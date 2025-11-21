package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github-stars-manager/utils"
	"go.uber.org/zap"
)

// FileRepository 基于文件系统的数据存储实现
type FileRepository struct {
	dataDir string
	mu      sync.RWMutex
	logger  *zap.Logger
	githubCli *utils.GithubUtil
}

// NewFileRepository 创建一个新的文件存储实例
func NewFileRepository(logger *zap.Logger, githubCli *utils.GithubUtil) Repository {
	// 确保数据目录存在
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Error("创建数据目录失败", zap.Error(err))
		panic(err)
	}

	return &FileRepository{
		dataDir: dataDir,
		logger: logger,
		githubCli: githubCli,
	}
}

// GetReposWithTag 获取带标签的仓库列表
func (f *FileRepository) GetReposWithTag() ([]utils.Repo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.logger.Debug("从文件系统获取带标签的仓库列表")
	data, err := os.ReadFile(filepath.Join(f.dataDir, "repos.json"))
	if err != nil {
		if os.IsNotExist(err) {
			f.logger.Debug("仓库数据文件不存在")
			return nil, err
		}
		f.logger.Error("读取仓库数据文件失败", zap.Error(err))
		return nil, err
	}

	var repos []utils.Repo
	err = json.Unmarshal(data, &repos)
	if err != nil {
		f.logger.Error("解析仓库数据失败", zap.Error(err))
		return nil, err
	}
	
	// 加载标签信息并合并到仓库数据中
	tags, err := f.loadTags()
	if err != nil {
		f.logger.Error("加载标签信息失败", zap.Error(err))
		return nil, err
	}
	
	// 将标签和分类信息附加到对应的仓库
	for i := range repos {
		if tagInfo, exists := tags[repos[i].ID]; exists {
			repos[i].Tag = tagInfo.Tag
			repos[i].Category = tagInfo.Category
		}
	}

	f.logger.Debug("成功从文件系统获取带标签的仓库列表", zap.Int("count", len(repos)))
	return repos, nil
}

// SaveRepos 保存仓库列表
func (f *FileRepository) SaveRepos(repos []utils.Repo) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.logger.Debug("保存仓库列表到文件系统", zap.Int("count", len(repos)))
	data, err := json.Marshal(repos)
	if err != nil {
		f.logger.Error("序列化仓库数据失败", zap.Error(err))
		return err
	}

	filename := filepath.Join(f.dataDir, "repos.json")
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		f.logger.Error("写入仓库数据文件失败", zap.Error(err))
		return err
	}

	f.logger.Debug("成功保存仓库列表到文件系统")
	return nil
}

// GetRepoTag 获取仓库标签信息
func (f *FileRepository) GetRepoTag(id int64) (*RepoTag, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.logger.Debug("从文件系统获取仓库标签信息", zap.Int64("id", id))
	tags, err := f.loadTags()
	if err != nil {
		return nil, err
	}

	tag, exists := tags[id]
	if !exists {
		f.logger.Debug("未找到仓库标签信息", zap.Int64("id", id))
		return nil, nil
	}

	f.logger.Debug("成功从文件系统获取仓库标签信息")
	return &tag, nil
}

// SaveRepoTag 保存仓库标签信息
func (f *FileRepository) SaveRepoTag(tag *RepoTag) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.logger.Debug("保存仓库标签信息到文件系统", zap.Int64("id", tag.ID))
	tags, err := f.loadTags()
	if err != nil {
		return err
	}

	tags[tag.ID] = *tag
	return f.saveTags(tags)
}

// DeleteRepoTag 删除仓库标签信息
func (f *FileRepository) DeleteRepoTag(id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.logger.Debug("从文件系统删除仓库标签信息", zap.Int64("id", id))
	tags, err := f.loadTags()
	if err != nil {
		return err
	}

	delete(tags, id)
	return f.saveTags(tags)
}

// GetStats 获取统计信息
func (f *FileRepository) GetStats() (*Stats, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.logger.Debug("从文件系统获取统计信息")
	repos, err := f.GetReposWithTag()
	if err != nil {
		f.logger.Error("获取仓库数据失败", zap.Error(err))
		return nil, err
	}

	stats := &Stats{}
	stats.TotalRepos = len(repos)

	// 计算已分析的仓库数量（有标签或分类的仓库）
	for _, repo := range repos {
		if repo.Tag != "" || repo.Category != "" {
			stats.AnalyzedRepos++
		}
	}
	
	// 获取上次同步时间
	stats.LastSync, _ = f.LoadSyncTime()

	f.logger.Debug("成功从文件系统获取统计信息", 
		zap.Int("total", stats.TotalRepos),
		zap.Int("analyzed", stats.AnalyzedRepos))
	return stats, nil
}

// SaveSyncTime 保存同步时间
func (f *FileRepository) SaveSyncTime() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.logger.Debug("保存同步时间到文件系统")
	now := time.Now().Format(time.RFC3339)
	filename := filepath.Join(f.dataDir, "last_sync.txt")
	err := os.WriteFile(filename, []byte(now), 0644)
	if err != nil {
		f.logger.Error("写入同步时间文件失败", zap.Error(err))
		return err
	}

	f.logger.Debug("成功保存同步时间到文件系统")
	return nil
}

// LoadSyncTime 加载同步时间
func (f *FileRepository) LoadSyncTime() (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.logger.Debug("从文件系统加载同步时间")
	filename := filepath.Join(f.dataDir, "last_sync.txt")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			f.logger.Debug("同步时间文件不存在")
			return "", nil
		}
		f.logger.Error("读取同步时间文件失败", zap.Error(err))
		return "", err
	}

	f.logger.Debug("成功从文件系统加载同步时间")
	return string(data), nil
}

// loadTags 加载所有标签信息
func (f *FileRepository) loadTags() (map[int64]RepoTag, error) {
	filename := filepath.Join(f.dataDir, "repo_tags.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[int64]RepoTag), nil
		}
		f.logger.Error("读取标签数据文件失败", zap.Error(err))
		return nil, err
	}

	var tags map[int64]RepoTag
	err = json.Unmarshal(data, &tags)
	if err != nil {
		f.logger.Error("解析标签数据失败", zap.Error(err))
		return nil, err
	}

	return tags, nil
}

// saveTags 保存所有标签信息
func (f *FileRepository) saveTags(tags map[int64]RepoTag) error {
	// 按ID排序，便于调试
	keys := make([]int64, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	filename := filepath.Join(f.dataDir, "repo_tags.json")
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		f.logger.Error("序列化标签数据失败", zap.Error(err))
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		f.logger.Error("写入标签数据文件失败", zap.Error(err))
		return err
	}

	return nil
}