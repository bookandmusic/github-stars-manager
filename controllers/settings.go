package controllers

import (
	"net/http"

	"github-stars-manager/config"
	"github-stars-manager/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SettingsHandler 处理设置相关的请求
type SettingsHandler struct {
	config      *config.Config
	logger      *zap.Logger
	openaiCli   *utils.OpenAIUtil // 添加OpenAI工具实例
	settingsCli *utils.SettingsUtil
}

// NewSettingsHandler 创建设置处理器实例
func NewSettingsHandler(config *config.Config, logger *zap.Logger, openaiCli *utils.OpenAIUtil, settingsCli *utils.SettingsUtil) *SettingsHandler {
	return &SettingsHandler{
		config:      config,
		logger:      logger,
		openaiCli:   openaiCli,
		settingsCli: settingsCli,
	}
}

// SettingsPage 展示设置页面
func (h *SettingsHandler) SettingsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "settings.html", nil)
}

// GetSettings 获取当前设置
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsCli.LoadSettings()
	if err != nil {
		h.logger.Error("加载设置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载设置失败"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// SaveSettings 保存设置
func (h *SettingsHandler) SaveSettings(c *gin.Context) {
	var settings utils.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		h.logger.Error("绑定JSON失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求数据格式错误"})
		return
	}
	h.settingsCli.SaveSettings(&settings)
	c.JSON(http.StatusOK, gin.H{"message": "设置已保存"})
}

// TestOpenAI 测试OpenAI连接
func (h *SettingsHandler) TestOpenAI(c *gin.Context) {
	var openaiConfig utils.OpenAISettings
	if err := c.ShouldBindJSON(&openaiConfig); err != nil {
		h.logger.Error("绑定JSON失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求数据格式错误"})
		return
	}

	// 调用OpenAI工具类测试连接
	_, err := h.openaiCli.CallWithPrompt(openaiConfig, "你好，请简单介绍一下你自己。")
	if err != nil {
		h.logger.Error("测试OpenAI连接失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "连接失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "连接成功"})
}

// TestWebDAV 测试WebDAV连接
func (h *SettingsHandler) TestWebDAV(c *gin.Context) {
	var webdavConfig utils.WebDAVSettings
	if err := c.ShouldBindJSON(&webdavConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 检查必要字段
	if webdavConfig.Url == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请提供WebDAV服务器地址"})
		return
	}

	// 发送一个PROPFIND请求测试连接
	client := &http.Client{}
	req, err := http.NewRequest("PROPFIND", webdavConfig.Url, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建请求失败: " + err.Error()})
		return
	}

	// 如果提供了用户名和密码，则添加基本认证
	if webdavConfig.Username != "" && webdavConfig.Password != "" {
		req.SetBasicAuth(webdavConfig.Username, webdavConfig.Password)
	}

	req.Header.Set("Depth", "0")
	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// WebDAV服务器通常会返回207 Multi-Status或200 OK表示成功
	if resp.StatusCode == http.StatusMultiStatus || resp.StatusCode == http.StatusOK {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "连接成功"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "连接失败，状态码: " + resp.Status})
	}
}
