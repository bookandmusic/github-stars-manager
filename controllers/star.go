package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github-stars-manager/config"
	"github-stars-manager/repository"
	"github-stars-manager/session"
	"github-stars-manager/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的连接
	},
}

// 为 WebSocket 连接添加写入锁
type SafeWebSocketConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (s *SafeWebSocketConn) WriteJSON(v interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteJSON(v)
}

type RepoData struct {
	Tag      string `json:"tag"`
	Category string `json:"category"`
}

// AIAnalysisResult AI分析结果
type AIAnalysisResult struct {
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

// GetRepos 获取仓库列表
func (h *StarHandler) GetRepos(c *gin.Context) {
	h.logger.Info("获取仓库列表")
	
	// 尝试从本地数据库加载带标签的仓库
	repos, err := h.repo.GetReposWithTag()
	if err == nil {
		// 尝试加载AI分析的描述信息
		for i := range repos {
			tagInfo, err := h.repo.GetRepoTag(repos[i].ID)
			if err == nil && tagInfo != nil && tagInfo.Description != "" {
				// 如果有AI分析的描述，则使用它替换原始描述
				repos[i].Description = tagInfo.Description
			}
		}
		
		h.logger.Info("成功从本地获取仓库列表", zap.Int("count", len(repos)))
		c.JSON(http.StatusOK, repos)
		return
	}
	
	// 如果无法加载本地数据，则从API获取
	h.logger.Warn("无法从本地加载仓库数据，从API获取", zap.Error(err))
	s, exists := c.Get("session")
	if !exists {
		h.logger.Error("会话不存在")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	sess := s.(*session.SessionData)
	repos, err = h.githubCli.GetStarredRepos(sess.AccessToken)
	if err != nil {
		h.logger.Error("从API获取仓库列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取仓库列表失败"})
		return
	}
	
	// 保存新获取的数据
	if saveErr := h.repo.SaveRepos(repos); saveErr != nil {
		h.logger.Warn("保存仓库数据失败", zap.Error(saveErr))
	}
	h.repo.SaveSyncTime()
	
	// 对于新获取的数据，暂时不填充AI描述（因为还没有分析）
	h.logger.Info("成功从API获取仓库列表", zap.Int("count", len(repos)))
	c.JSON(http.StatusOK, repos)
}

// AnalyzeRepo 使用AI分析仓库
func (h *StarHandler) AnalyzeRepo(c *gin.Context) {
	repoID := c.Param("id")
	h.logger.Info("开始分析仓库", zap.String("repo_id", repoID))
	
	// 获取仓库ID
	id, err := strconv.ParseInt(repoID, 10, 64)
	if err != nil {
		h.logger.Error("仓库ID格式错误", zap.String("repo_id", repoID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "仓库ID格式错误"})
		return
	}
	
	// 获取仓库信息
	repo, err := h.getRepoByID(id)
	if err != nil {
		h.logger.Error("获取仓库信息失败", zap.Int64("repo_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定仓库"})
		return
	}
	
	// 获取AI配置
	settings, err := h.settingsCli.LoadSettings()
	if err != nil {
		h.logger.Error("加载AI设置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载AI设置失败"})
		return
	}
	openSetting := settings.OpenAI
	if openSetting.Key == "" || openSetting.Endpoint == "" || openSetting.Model == "" {
		h.logger.Warn("AI配置不完整")
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI配置不完整，请先在设置中配置AI参数"})
		return
	}
	
	// 获取仓库README内容
	readmeContent, err := h.getRepoReadmeWithToken(repo.HTMLURL, c)
	if err != nil {
		h.logger.Warn("获取仓库README失败", zap.Error(err))
	}
	
	// 构造AI分析提示
	prompt := h.buildAIAnalysisPrompt(repo, readmeContent)
	
	// 调用AI分析
	analysisResult, err := h.callAIAnalysis(openSetting, prompt)
	if err != nil {
		h.logger.Error("AI分析失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI分析失败: " + err.Error()})
		return
	}
	
	// 验证分析结果
	if analysisResult.Category == "" || len(analysisResult.Tags) == 0 {
		h.logger.Error("AI分析结果不完整", 
			zap.String("category", analysisResult.Category),
			zap.Strings("tags", analysisResult.Tags),
			zap.String("description", analysisResult.Description))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI分析结果不完整，请稍后重试"})
		return
	}
	
	// 保存分析结果
	err = h.saveAnalysisResult(id, analysisResult)
	if err != nil {
		h.logger.Error("保存分析结果失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存分析结果失败"})
		return
	}
	
	// 返回AI分析结果
	c.JSON(http.StatusOK, analysisResult)
}

// getRepoByID 根据ID获取仓库信息
func (h *StarHandler) getRepoByID(repoID int64) (*utils.Repo, error) {
	repos, err := h.repo.GetReposWithTag()
	if err != nil {
		return nil, err
	}
	
	for _, repo := range repos {
		if repo.ID == repoID {
			return &repo, nil
		}
	}
	
	return nil, fmt.Errorf("未找到ID为%d的仓库", repoID)
}

// getRepoReadmeWithToken 使用访问令牌获取仓库README内容
func (h *StarHandler) getRepoReadmeWithToken(repoURL string,  c *gin.Context) (string, error) {
	// 从上下文中获取session
	s, exists := c.Get("session")
	if !exists {
		return "", fmt.Errorf("无法获取用户会话")
	}
	
	sess := s.(*session.SessionData)
	if sess.AccessToken == "" {
		return "", fmt.Errorf("访问令牌为空")
	}
	
	// 从仓库URL提取用户名和仓库名
	// 例如: https://github.com/user/repo -> user/repo
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效的仓库URL: %s", repoURL)
	}
	
	repoFullName := parts[len(parts)-2] + "/" + parts[len(parts)-1]
	
	// 构造API URL获取README
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/readme", repoFullName)
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Authorization", "token "+sess.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求README失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		// 如果无法获取README，不视为错误，只记录警告
		h.logger.Warn("获取仓库README失败", 
			zap.String("repo", repoFullName), 
			zap.Int("status", resp.StatusCode))
		return "", nil
	}
	
	var readmeResp struct {
		Content string `json:"content"`
		Encoding string `json:"encoding"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&readmeResp)
	if err != nil {
		return "", fmt.Errorf("解析README响应失败: %w", err)
	}
	
	// 如果内容是base64编码的，需要解码
	if readmeResp.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(readmeResp.Content)
		if err != nil {
			h.logger.Warn("解码README内容失败", zap.Error(err))
			return "", nil
		}
		return string(decoded), nil
	}
	
	return readmeResp.Content, nil
}

// buildAIAnalysisPrompt 构造AI分析提示
func (h *StarHandler) buildAIAnalysisPrompt(repo *utils.Repo, readme string) string {
	topics := "无"
	if len(repo.Topics) > 0 {
		topics = strings.Join(repo.Topics, ", ")
	}
	
	languages := "无"
	if len(repo.Languages) > 0 {
		languages = strings.Join(repo.Languages, ", ")
	}
	
	description := "无描述"
	if repo.Description != "" {
		description = repo.Description
	}
	
	readmeText := "无README"
	if readme != "" {
		// 限制README长度以避免超出token限制
		if len(readme) > 2000 {
			readme = readme[:2000] + "...(内容过长已截断)"
		}
		readmeText = readme
	}
	
	// 定义标准分类列表（使用中文）
	categories := []string{
		"前端",
		"后端", 
		"移动开发",
		"工具",
		"数据库",
		"运维",
		"人工智能",
		"安全",
		"物联网",
		"游戏",
	}
	
	categoryList := ""
	for _, cat := range categories {
		categoryList += fmt.Sprintf("- %s\n", cat)
	}
	
	prompt := fmt.Sprintf(`你是一个专业的GitHub项目分析师，请分析以下GitHub仓库信息，并用中文提供结构化的分析结果。

仓库信息：
- 名称：%s
- 描述：%s
- 主要编程语言：%s
- 编程语言列表：%s
- 主题标签：%s
- README内容：%s

请根据以上信息提供以下三方面的分析：

1. 分类：请从以下分类中选择最合适的一个分类：
%s

2. 标签：请提供3个最能代表此仓库的标签，尽量使用中文，用逗号分隔

3. 描述：请用中文写一段简洁明了的描述（不超过100字），这个描述应该比原始描述更详细和准确

请严格按照以下JSON格式返回结果，不要包含其他内容：
{
  "category": "分类名称（如前端、后端等）",
  "tags": ["标签1", "标签2", "标签3"],
  "description": "项目描述"
}`, 
		repo.Name, 
		description,
		repo.Language, 
		languages, 
		topics, 
		readmeText,
		categoryList)
	
	return prompt
}

// callAIAnalysis 调用AI进行分析
func (h *StarHandler) callAIAnalysis(settings utils.OpenAISettings, prompt string) (*AIAnalysisResult, error) {
	// 调用OpenAI工具类
	content, err := h.openaiCli.CallWithPrompt(settings, prompt)
	if err != nil {
		h.logger.Error("调用OpenAI API失败", zap.Error(err))
		return nil, fmt.Errorf("调用OpenAI API失败: %w", err)
	}

	var result AIAnalysisResult
	// 尝试直接解析JSON
	err = json.Unmarshal([]byte(content), &result)
	if err != nil {
		// 如果直接解析失败，尝试提取其中的JSON
		h.logger.Debug("直接解析JSON失败，尝试提取JSON", zap.String("content", content))

		// 查找第一个 { 和最后一个 } 之间的内容
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")

		if start >= 0 && end > start {
			jsonStr := content[start : end+1]
			// 清理JSON字符串中的特殊字符
			jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
			jsonStr = strings.ReplaceAll(jsonStr, "\\", "")

			h.logger.Debug("提取的JSON字符串", zap.String("json", jsonStr))

			// 先尝试解析为map，处理tags可能为字符串的情况
			var rawResult map[string]interface{}
			err = json.Unmarshal([]byte(jsonStr), &rawResult)
			if err != nil {
				h.logger.Error("解析提取的JSON失败",
					zap.Error(err),
					zap.String("json", jsonStr))
				return nil, fmt.Errorf("解析AI返回的JSON失败: %w", err)
			}

			// 处理category
			if category, ok := rawResult["category"].(string); ok {
				result.Category = category
			}

			// 处理description
			if description, ok := rawResult["description"].(string); ok {
				result.Description = description
			}

			// 处理tags，可能是字符串或字符串数组
			if tags, ok := rawResult["tags"]; ok {
				switch v := tags.(type) {
				case string:
					// 如果tags是字符串，按逗号分割
					result.Tags = strings.Split(v, ",")
					// 清理每个标签的空格
					for i, tag := range result.Tags {
						result.Tags[i] = strings.TrimSpace(tag)
					}
				case []interface{}:
					// 如果tags是数组，转换为字符串数组
					result.Tags = make([]string, len(v))
					for i, tag := range v {
						result.Tags[i] = fmt.Sprintf("%v", tag)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("AI未返回有效的JSON格式结果: %s", content)
		}
	}

	// 验证结果
	if result.Category == "" {
		return nil, fmt.Errorf("AI返回的分类为空")
	}

	if len(result.Tags) == 0 {
		return nil, fmt.Errorf("AI返回的标签为空")
	}

	// 限制标签数量为3个
	if len(result.Tags) > 3 {
		result.Tags = result.Tags[:3]
	}

	// 清理标签中的空格
	for i, tag := range result.Tags {
		result.Tags[i] = strings.TrimSpace(tag)
	}

	h.logger.Debug("AI分析成功",
		zap.String("category", result.Category),
		zap.Strings("tags", result.Tags),
		zap.String("description", result.Description))

	return &result, nil
}

// saveAnalysisResult 保存分析结果
func (h *StarHandler) saveAnalysisResult(repoID int64, result *AIAnalysisResult) error {
	// 创建RepoTag对象
	repoTag := &repository.RepoTag{
		ID:          repoID,
		Category:    result.Category,
		Tag:         strings.Join(result.Tags, ","),
		Description: result.Description,
	}
	
	h.logger.Debug("保存AI分析结果",
		zap.Int64("repo_id", repoID),
		zap.String("category", result.Category),
		zap.Strings("tags", result.Tags),
		zap.String("description", result.Description))
	
	// 保存到仓库（使用现有的SaveRepoTag方法）
	err := h.repo.SaveRepoTag(repoTag)
	if err != nil {
		h.logger.Error("保存仓库标签信息失败",
			zap.Int64("repo_id", repoID),
			zap.Error(err))
		return fmt.Errorf("保存仓库标签信息失败: %w", err)
	}
	
	h.logger.Info("AI分析结果保存成功", zap.Int64("repo_id", repoID))
	return nil
}

type SyncProgress struct {
	Type     string `json:"type"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
	Total    int    `json:"total,omitempty"`
	Current  int    `json:"current,omitempty"`
}

// StarHandler 处理star相关功能的结构体
type StarHandler struct {
	repo   repository.Repository
	logger *zap.Logger
	config *config.Config
	openaiCli *utils.OpenAIUtil
	settingsCli *utils.SettingsUtil
	githubCli *utils.GithubUtil
}

// NewStarHandler 创建一个新的StarHandler实例
func NewStarHandler(
	repo repository.Repository, 
	logger *zap.Logger, 
	config *config.Config, 
	openaiCli *utils.OpenAIUtil, 
	settingsCli *utils.SettingsUtil,
	githubCli *utils.GithubUtil,
	) *StarHandler {
	return &StarHandler{
		repo:   repo,
		logger: logger,
		config: config,
		openaiCli: openaiCli,
		settingsCli: settingsCli,
		githubCli: githubCli,
	}
}

// IndexPage 首页处理器
func (h *StarHandler) IndexPage(c *gin.Context) {
	h.logger.Info("访问首页")
	c.HTML(http.StatusOK, "index.html", nil)
}

// GetUser 获取用户信息
func (h *StarHandler) GetUser(c *gin.Context) {
	h.logger.Info("获取用户信息")
	s, _ := c.Get("session")
	sess := s.(*session.SessionData)
	c.JSON(http.StatusOK, gin.H{
		"login":      sess.UserName,
		"avatar_url": sess.AvatarURL,
	})
}

// GetStats 获取统计信息
func (h *StarHandler) GetStats(c *gin.Context) {
	h.logger.Info("获取统计信息")
	s, _ := c.Get("session")
	sess := s.(*session.SessionData)
	
	stats, err := h.repo.GetStats()
	if err != nil {
		// 如果获取统计数据失败，重新计算
		h.logger.Warn("获取统计数据失败，重新计算", zap.Error(err))
		repos, err := h.repo.GetReposWithTag()
		if err != nil {
			// 如果本地没有数据，则从API获取
			h.logger.Warn("无法从本地加载仓库数据，从API获取", zap.Error(err))
			repos, _ = h.githubCli.GetStarredRepos(sess.AccessToken)
		}
		
		stats = &repository.Stats{
			TotalRepos:    len(repos),
			AnalyzedRepos: 0,
		}
		
		// 计算已分析的仓库数量（有标签或分类的仓库）
		for _, repo := range repos {
			if repo.Tag != "" || repo.Category != "" {
				stats.AnalyzedRepos++
			}
		}
		
		// 获取上次同步时间
		stats.LastSync, _ = h.repo.LoadSyncTime()
		
		h.repo.SaveRepos(repos)
		h.repo.SaveSyncTime()
	}
	
	h.logger.Info("成功获取统计信息", 
		zap.Int("total_repos", stats.TotalRepos),
		zap.Int("analyzed_repos", stats.AnalyzedRepos))
	c.JSON(http.StatusOK, stats)
}

// SyncStars 同步stars
func (h *StarHandler) SyncStars(c *gin.Context) {
	h.logger.Info("开始同步stars")
	s, _ := c.Get("session")
	sess := s.(*session.SessionData)
	githubRepos, err := h.githubCli.GetStarredRepos(sess.AccessToken)
	if err != nil {
		h.logger.Error("同步仓库失败", zap.Error(err))
		c.JSON(500, gin.H{"msg": "同步失败", "error": err.Error()})
		return
	}
	
	// 加载本地已有的仓库数据
	localRepos, err := h.repo.GetReposWithTag()
	if err != nil {
		// 如果没有本地数据，则全部使用github数据
		h.logger.Warn("无法加载本地仓库数据", zap.Error(err))
		localRepos = []utils.Repo{}
	}
	
	// 创建一个map来快速查找本地仓库
	localRepoMap := make(map[int64]utils.Repo)
	for _, repo := range localRepos {
		localRepoMap[repo.ID] = repo
	}
	
	// 合并数据：保留本地编辑的信息，添加新的仓库，移除不存在的仓库
	var mergedRepos []utils.Repo
	githubRepoMap := make(map[int64]bool)
	
	for _, githubRepo := range githubRepos {
		githubRepoMap[githubRepo.ID] = true
		
		// 检查该仓库是否已经在本地存在
		if localRepo, exists := localRepoMap[githubRepo.ID]; exists {
			// 如果存在，保留用户编辑的信息
			mergedRepos = append(mergedRepos, localRepo)
		} else {
			// 如果不存在，添加新的仓库（设置默认值）
			githubRepo.Tag = ""
			githubRepo.Category = ""
			mergedRepos = append(mergedRepos, githubRepo)
		}
	}
	
	// 保存合并后的仓库数据和同步时间
	err = h.repo.SaveRepos(mergedRepos)
	if err != nil {
		h.logger.Error("保存仓库数据失败", zap.Error(err))
		c.JSON(500, gin.H{"msg": "保存数据失败", "error": err.Error()})
		return
	}
	
	err = h.repo.SaveSyncTime()
	if err != nil {
		h.logger.Error("保存同步时间失败", zap.Error(err))
	}
	
	h.logger.Info("同步完成", zap.Int("count", len(mergedRepos)))
	c.JSON(http.StatusOK, gin.H{
		"msg": "同步完成",
		"count": len(mergedRepos),
	})
}

// SyncProgressWS 同步进度WebSocket
func (h *StarHandler) SyncProgressWS(c *gin.Context) {
	h.logger.Info("开始WebSocket同步进度")
	// 升级到 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("无法升级到WebSocket连接", zap.Error(err))
		c.JSON(500, gin.H{"error": "无法升级到 WebSocket 连接"})
		return
	}
	defer conn.Close()

	// 创建线程安全的 WebSocket 连接包装器
	safeConn := &SafeWebSocketConn{conn: conn}

	// 获取会话信息
	s, exists := c.Get("session")
	if !exists {
		h.logger.Error("未找到会话信息")
		safeConn.WriteJSON(SyncProgress{
			Type:    "error",
			Message: "未找到会话信息",
		})
		return
	}
	sess := s.(*session.SessionData)

	// 发送初始消息
	safeConn.WriteJSON(SyncProgress{
		Type:     "start",
		Message:  "开始同步仓库数据",
		Progress: 0,
	})

	// 获取 GitHub 星标仓库
	githubRepos, err := func() ([]utils.Repo, error) {
		safeConn.WriteJSON(SyncProgress{
			Type:     "info",
			Message:  "正在获取星标仓库列表...",
			Progress: 5,
		})

		// 先获取仓库总数
		totalCount, err := h.getTotalStarredRepos(sess.AccessToken)
		if err != nil {
			return nil, err
		}

		safeConn.WriteJSON(SyncProgress{
			Type:     "info",
			Message:  fmt.Sprintf("找到 %d 个星标仓库，正在获取详细信息...", totalCount),
			Progress: 10,
			Total:    totalCount,
		})

		// 获取所有仓库的详细信息（并发）
		repos, err := h.getAllStarredReposDetailed(sess.AccessToken, totalCount, func(processed int) {
			// 计算进度 (10-80% 范围)
			progress := 10 + int(float64(processed)/float64(totalCount)*70)
			if progress > 80 {
				progress = 80
			}
			
			safeConn.WriteJSON(SyncProgress{
				Type:     "progress",
				Message:  fmt.Sprintf("正在获取仓库详细信息 (%d/%d)", processed, totalCount),
				Progress: progress,
				Current:  processed,
				Total:    totalCount,
			})
		})
		
		if err != nil {
			return nil, err
		}

		safeConn.WriteJSON(SyncProgress{
			Type:     "info",
			Message:  fmt.Sprintf("已完成获取仓库信息，共 %d 个仓库", len(repos)),
			Progress: 80,
			Total:    len(repos),
		})

		return repos, nil
	}()

	if err != nil {
		h.logger.Error("获取GitHub仓库失败", zap.Error(err))
		safeConn.WriteJSON(SyncProgress{
			Type:    "error",
			Message: "获取 GitHub 仓库失败: " + err.Error(),
		})
		return
	}

	totalRepos := len(githubRepos)

	// 加载本地已有的仓库数据
	localRepos, err := h.repo.GetReposWithTag()
	if err != nil {
		// 如果没有本地数据，则全部使用github数据
		h.logger.Warn("无法加载本地仓库数据", zap.Error(err))
		localRepos = []utils.Repo{}
	}

	safeConn.WriteJSON(SyncProgress{
		Type:     "info",
		Message:  "加载本地仓库数据",
		Progress: 85,
	})

	// 创建一个map来快速查找本地仓库
	localRepoMap := make(map[int64]utils.Repo)
	for _, repo := range localRepos {
		localRepoMap[repo.ID] = repo
	}

	// 合并数据：保留本地编辑的信息，添加新的仓库，移除不存在的仓库
	var mergedRepos []utils.Repo
	githubRepoMap := make(map[int64]bool)

	safeConn.WriteJSON(SyncProgress{
		Type:     "info",
		Message:  "处理仓库数据",
		Progress: 90,
	})

	for i, githubRepo := range githubRepos {
		githubRepoMap[githubRepo.ID] = true

		// 检查该仓库是否已经在本地存在
		if _, exists := localRepoMap[githubRepo.ID]; exists {
			// 如果存在，保留用户编辑的信息
			mergedRepos = append(mergedRepos, githubRepo)
		} else {
			// 如果不存在，添加新的仓库（设置默认值）
			githubRepo.Tag = ""
			githubRepo.Category = ""
			mergedRepos = append(mergedRepos, githubRepo)
		}

		// 每处理10%的仓库发送一次进度更新
		progress := 90 + int(float64(i+1)/float64(totalRepos)*5)
		if progress > 95 {
			progress = 95
		}
		if (i+1)%(totalRepos/10+1) == 0 || i == totalRepos-1 {
			safeConn.WriteJSON(SyncProgress{
				Type:     "progress",
				Message:  "处理仓库 " + strconv.Itoa(i+1) + "/" + strconv.Itoa(totalRepos),
				Progress: progress,
				Current:  i + 1,
				Total:    totalRepos,
			})
		}
	}

	safeConn.WriteJSON(SyncProgress{
		Type:     "info",
		Message:  "保存数据",
		Progress: 95,
	})

	// 保存合并后的仓库数据和同步时间
	err = h.repo.SaveRepos(mergedRepos)
	if err != nil {
		h.logger.Error("保存仓库数据失败", zap.Error(err))
	}
	
	err = h.repo.SaveSyncTime()
	if err != nil {
		h.logger.Error("保存同步时间失败", zap.Error(err))
	}

	safeConn.WriteJSON(SyncProgress{
		Type:     "complete",
		Message:  "同步完成，共处理 " + strconv.Itoa(len(mergedRepos)) + " 个仓库",
		Progress: 100,
		Total:    len(mergedRepos),
	})
	
	h.logger.Info("WebSocket同步完成", zap.Int("count", len(mergedRepos)))
}

// UpdateTag 更新标签
func (h *StarHandler) UpdateTag(c *gin.Context) {
	h.logger.Info("更新标签")
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var body struct {
		Tag string `json:"tag"`
	}
	if err := c.BindJSON(&body); err != nil {
		h.logger.Error("参数错误", zap.Error(err))
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	// 获取现有的标签信息
	tagInfo, err := h.repo.GetRepoTag(id)
	if err != nil || tagInfo == nil {
		tagInfo = &repository.RepoTag{ID: id}
	}
	
	// 更新标签
	tagInfo.Tag = body.Tag
	
	// 如果标签和分类都为空，则删除该记录
	if tagInfo.Tag == "" && tagInfo.Category == "" {
		err = h.repo.DeleteRepoTag(id)
		if err != nil {
			h.logger.Error("删除仓库标签失败", zap.Error(err))
		}
	} else {
		err = h.repo.SaveRepoTag(tagInfo)
		if err != nil {
			h.logger.Error("保存仓库标签失败", zap.Error(err))
			c.JSON(500, gin.H{"msg": "保存标签失败"})
			return
		}
	}
	
	h.logger.Info("标签更新成功", zap.Int64("repo_id", id))
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// UpdateCategory 更新分类
func (h *StarHandler) UpdateCategory(c *gin.Context) {
	h.logger.Info("更新分类")
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var body struct {
		Category string `json:"category"`
	}
	if err := c.BindJSON(&body); err != nil {
		h.logger.Error("参数错误", zap.Error(err))
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	// 获取现有的标签信息
	tagInfo, err := h.repo.GetRepoTag(id)
	if err != nil || tagInfo == nil {
		tagInfo = &repository.RepoTag{ID: id}
	}
	
	// 更新分类
	tagInfo.Category = body.Category
	
	// 如果标签和分类都为空，则删除该记录
	if tagInfo.Tag == "" && tagInfo.Category == "" {
		err = h.repo.DeleteRepoTag(id)
		if err != nil {
			h.logger.Error("删除仓库标签失败", zap.Error(err))
		}
	} else {
		err = h.repo.SaveRepoTag(tagInfo)
		if err != nil {
			h.logger.Error("保存仓库标签失败", zap.Error(err))
			c.JSON(500, gin.H{"msg": "保存分类失败"})
			return
		}
	}
	
	h.logger.Info("分类更新成功", zap.Int64("repo_id", id))
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// UpdateDescription 更新描述
func (h *StarHandler) UpdateDescription(c *gin.Context) {
	h.logger.Info("更新描述")
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var body struct {
		Description string `json:"description"`
	}
	if err := c.BindJSON(&body); err != nil {
		h.logger.Error("参数错误", zap.Error(err))
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	// 获取现有的标签信息
	tagInfo, err := h.repo.GetRepoTag(id)
	if err != nil || tagInfo == nil {
		tagInfo = &repository.RepoTag{ID: id}
	}
	
	// 更新描述
	tagInfo.Description = body.Description
	
	// 如果标签、分类和描述都为空，则删除该记录
	if tagInfo.Tag == "" && tagInfo.Category == "" && tagInfo.Description == "" {
		err = h.repo.DeleteRepoTag(id)
		if err != nil {
			h.logger.Error("删除仓库标签失败", zap.Error(err))
		}
	} else {
		err = h.repo.SaveRepoTag(tagInfo)
		if err != nil {
			h.logger.Error("保存仓库标签失败", zap.Error(err))
			c.JSON(500, gin.H{"msg": "保存描述失败"})
			return
		}
	}
	
	h.logger.Info("描述更新成功", zap.Int64("repo_id", id))
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// GetCategories 获取分类列表
func (h *StarHandler) GetCategories(c *gin.Context) {
	h.logger.Info("获取分类列表")
	categories := []map[string]string{
		{"value": "前端", "label": "前端"},
		{"value": "后端", "label": "后端"},
		{"value": "移动开发", "label": "移动开发"},
		{"value": "工具", "label": "工具"},
		{"value": "数据库", "label": "数据库"},
		{"value": "运维", "label": "运维"},
		{"value": "人工智能", "label": "人工智能"},
		{"value": "安全", "label": "安全"},
		{"value": "物联网", "label": "物联网"},
		{"value": "游戏", "label": "游戏"},
	}
	c.JSON(http.StatusOK, categories)
}

// 获取总的星标仓库数量
func (h *StarHandler) getTotalStarredRepos(accessToken string) (int, error) {
	h.logger.Debug("获取星标仓库总数")
	client := &http.Client{Timeout: 30 * time.Second}
	page := 1
	totalCount := 0
	
	// 循环遍历所有页面，计算总数量
	for {
		var resp *http.Response
		var err error
		
		// 添加重试机制
		for attempts := 0; attempts < 3; attempts++ {
			req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/user/starred?page=%d&per_page=100", page), nil)
			req.Header.Set("Authorization", "token "+accessToken)
			req.Header.Set("Accept", "application/vnd.github.v3+json")
			
			resp, err = client.Do(req)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(attempts+1) * time.Second) // 逐步增加等待时间
		}
		
		if err != nil {
			h.logger.Error("获取仓库列表失败", 
				zap.Error(err), 
				zap.Int("page", page))
			return 0, fmt.Errorf("获取第 %d 页仓库列表失败: %w", page, err)
		}
		
		var repos []interface{}
		err = json.NewDecoder(resp.Body).Decode(&repos)
		resp.Body.Close()
		
		if err != nil {
			h.logger.Error("解析仓库列表失败", 
				zap.Error(err), 
				zap.Int("page", page))
			return 0, fmt.Errorf("解析第 %d 页仓库列表失败: %w", page, err)
		}
		
		// 累加当前页面的仓库数量
		totalCount += len(repos)
		
		// 如果当前页面的仓库数量少于100，说明已经是最后一页
		if len(repos) < 100 {
			break
		}
		
		page++
	}
	
	h.logger.Debug("获取星标仓库总数完成", zap.Int("total", totalCount))
	return totalCount, nil
}

// 并发获取所有星标仓库的详细信息
func (h *StarHandler) getAllStarredReposDetailed(accessToken string, totalCount int, progressCallback func(int)) ([]utils.Repo, error) {
	h.logger.Debug("并发获取所有星标仓库详细信息")
	var allRepos []utils.Repo
	page := 1
	
	// 计算需要获取的页数
	totalPages := (totalCount + 99) / 100 // 每页100个，向上取整
	
	// 创建带缓存的通道用于接收结果
	resultChan := make(chan []utils.Repo, totalPages)
	errorChan := make(chan error, totalPages)
	processedCount := 0
	mu := sync.Mutex{}
	
	// 并发获取所有页面
	for page <= totalPages {
		go func(p int) {
			repos, err := h.getPageStarredRepos(accessToken, p, func(count int) {
				// 更新处理计数并调用进度回调
				mu.Lock()
				processedCount += count
				currentProcessed := processedCount
				mu.Unlock()
				
				progressCallback(currentProcessed)
			})
			if err != nil {
				h.logger.Error("获取页面仓库失败", 
					zap.Error(err), 
					zap.Int("page", p))
				errorChan <- err
				return
			}
			resultChan <- repos
		}(page)
		
		page++
	}
	
	// 收集所有结果
	for i := 0; i < totalPages; i++ {
		select {
		case repos := <-resultChan:
			mu.Lock()
			allRepos = append(allRepos, repos...)
			mu.Unlock()
			
		case err := <-errorChan:
			h.logger.Error("获取仓库详细信息失败", zap.Error(err))
			return nil, err
		}
	}
	
	h.logger.Debug("并发获取所有星标仓库详细信息完成", zap.Int("count", len(allRepos)))
	return allRepos, nil
}

// 获取单页星标仓库列表
func (h *StarHandler) getPageStarredRepos(accessToken string, page int, progressCallback func(int)) ([]utils.Repo, error) {
	h.logger.Debug("获取单页星标仓库列表", zap.Int("page", page))
	client := &http.Client{Timeout: 30 * time.Second}
	
	var resp *http.Response
	var err error
	
	// 添加重试机制
	for attempts := 0; attempts < 3; attempts++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/user/starred?page=%d&per_page=100", page), nil)
		req.Header.Set("Authorization", "token "+accessToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempts+1) * time.Second) // 逐步增加等待时间
	}
	
	if err != nil {
		h.logger.Error("获取页面仓库列表失败", 
			zap.Error(err), 
			zap.Int("page", page))
		return nil, fmt.Errorf("获取第 %d 页仓库列表失败: %w", page, err)
	}
	
	defer resp.Body.Close()
	
	var repos []struct {
		ID              int64  `json:"id"`
		Name            string `json:"name"`
		HTMLURL         string `json:"html_url"`
		StargazersCount int    `json:"stargazers_count"`
		Description     string `json:"description"`
		Language        string `json:"language"`
		Topics          []string `json:"topics"` // 添加Topics字段
	}
	err = json.NewDecoder(resp.Body).Decode(&repos)
	if err != nil {
		h.logger.Error("解析页面仓库列表失败", 
			zap.Error(err), 
			zap.Int("page", page))
		return nil, fmt.Errorf("解析第 %d 页仓库列表失败: %w", page, err)
	}
	
	var detailedRepos []utils.Repo
	for _, repo := range repos {
		// 构建仓库全名 (owner/repo)
		urlParts := strings.Split(strings.TrimSuffix(repo.HTMLURL, ".git"), "/")
		if len(urlParts) >= 2 {
			repoFullName := urlParts[len(urlParts)-2] + "/" + urlParts[len(urlParts)-1]
			
			// 获取仓库详细信息，添加重试机制
			var detailedRepo *utils.Repo
			var repoErr error
			for attempts := 0; attempts < 3; attempts++ {
				detailedRepo, repoErr = h.githubCli.GetRepoDetails(accessToken, repoFullName)
				if repoErr == nil && detailedRepo != nil {
					break
				}
				time.Sleep(time.Duration(attempts+1) * time.Second)
			}
			
			if repoErr == nil && detailedRepo != nil {
				detailedRepo.Tag = ""      // 初始化用户标签为空
				detailedRepo.Category = ""  // 初始化分类为空
				detailedRepos = append(detailedRepos, *detailedRepo)
			} else {
				// 如果获取详细信息失败，使用基础信息
				h.logger.Warn("获取仓库详细信息失败，使用基础信息", 
					zap.Error(repoErr),
					zap.String("repo", repoFullName))
				basicRepo := utils.Repo{
					ID:              repo.ID,
					Name:            repo.Name,
					HTMLURL:         repo.HTMLURL,
					StargazersCount: repo.StargazersCount,
					Description:     repo.Description,
					Language:        repo.Language,
					Languages:       []string{},
					Topics:          repo.Topics,
					Tag:             "",
					Category:        "",
				}
				detailedRepos = append(detailedRepos, basicRepo)
			}
		} else {
			// 如果无法解析仓库全名，使用基础信息
			h.logger.Warn("无法解析仓库全名，使用基础信息", 
				zap.String("url", repo.HTMLURL))
			basicRepo := utils.Repo{
				ID:              repo.ID,
				Name:            repo.Name,
				HTMLURL:         repo.HTMLURL,
				StargazersCount: repo.StargazersCount,
				Description:     repo.Description,
				Language:        repo.Language,
				Languages:       []string{},
				Topics:          repo.Topics,
				Tag:             "",
				Category:        "",
			}
			detailedRepos = append(detailedRepos, basicRepo)
		}
		
		// 每处理一个仓库就调用进度回调
		if progressCallback != nil {
			progressCallback(1)
		}
	}
	
	h.logger.Debug("获取单页星标仓库列表完成", 
		zap.Int("page", page), 
		zap.Int("count", len(detailedRepos)))
	return detailedRepos, nil
}