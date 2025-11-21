package controllers

import (
	"fmt"
	"net/http"

	"github-stars-manager/config"
	"github-stars-manager/session"
	"github-stars-manager/utils"
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

// AuthHandler 处理认证相关功能的结构体
type AuthHandler struct {
	config *config.Config
	logger *zap.Logger
	githubCli *utils.GithubUtil
}

// NewAuthHandler 创建一个新的AuthHandler实例
func NewAuthHandler(config *config.Config, logger *zap.Logger, githubCli *utils.GithubUtil) *AuthHandler {
	return &AuthHandler{
		config: config,
		logger: logger,
		githubCli: githubCli,
	}
}

// LoginPage 登录页面
func (h *AuthHandler) LoginPage(c *gin.Context) {
	h.logger.Info("访问登录页面")
	c.HTML(http.StatusOK, "login.html", nil)
}

// TokenLogin 使用token登录
func (h *AuthHandler) TokenLogin(c *gin.Context) {
	h.logger.Info("使用token登录")
	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.logger.Error("绑定JSON失败", zap.Error(err))
		c.JSON(400, gin.H{"msg": "JSON格式错误", "error": err.Error()})
		return
	}

	// 检查token是否为空
	if body.Token == "" {
		h.logger.Warn("token为空")
		c.JSON(400, gin.H{"msg": "token不能为空"})
		return
	}

	// 验证token
	user, err := h.githubCli.GetUserInfo(body.Token)
	if err != nil {
		h.logger.Error("获取用户信息失败", zap.Error(err))
		c.JSON(401, gin.H{"msg": "token无效"})
		return
	}

	// 创建session
	sess := session.NewSessionData()
	sess.AccessToken = body.Token
	sess.UserName = user.Login
	sess.AvatarURL = user.AvatarURL
	session.SetSession(c, sess)

	h.logger.Info("token登录成功", zap.String("user", user.Login))
	c.JSON(http.StatusOK, gin.H{"msg": "登录成功"})
}

// GitHubLogin GitHub登录
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	h.logger.Info("GitHub登录")
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user,user:email,repo",
		h.config.GitHubClientID,
		h.config.RedirectURL,
	)
	c.Redirect(http.StatusFound, redirectURL)
}

// GitHubCallback GitHub登录回调
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	h.logger.Info("GitHub登录回调")
	code := c.Query("code")
	if code == "" {
		h.logger.Error("缺少code参数")
		c.JSON(400, gin.H{"msg": "缺少code参数"})
		return
	}

	// 获取access token
	token, err := h.githubCli.GetAccessToken(h.config.GitHubClientID, h.config.GitHubClientSecret, code)
	if err != nil {
		h.logger.Error("获取access token失败", zap.Error(err))
		c.JSON(500, gin.H{"msg": "获取access token失败"})
		return
	}

	// 获取用户信息
	user, err := h.githubCli.GetUserInfo(token)
	if err != nil {
		h.logger.Error("获取用户信息失败", zap.Error(err))
		c.JSON(500, gin.H{"msg": "获取用户信息失败"})
		return
	}

	// 创建session
	sess := session.NewSessionData()
	sess.AccessToken = token
	sess.UserName = user.Login
	sess.AvatarURL = user.AvatarURL
	session.SetSession(c, sess)

	h.logger.Info("GitHub登录成功", zap.String("user", user.Login))
	c.Redirect(http.StatusFound, "/")
}

// AuthMiddleware 认证中间件
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, err := session.GetSession(c)
		if err != nil || sess == nil {
			h.logger.Warn("未登录访问受保护资源", zap.String("path", c.Request.URL.Path))
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 将session信息存储到context中
		c.Set("session", sess)
		c.Next()
	}
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	h.logger.Info("用户登出")
	session.ClearSession(c)
	c.Redirect(http.StatusFound, "/login")
}