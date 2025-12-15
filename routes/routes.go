package routes

import (
	"github-stars-manager/config"
	"github-stars-manager/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server 封装gin引擎和配置
type Server struct {
	Engine *gin.Engine
	Config *config.Config
}

// NewServer 创建一个新的服务器实例
func NewServer(engine *gin.Engine, config *config.Config) *Server {
	return &Server{
		Engine: engine,
		Config: config,
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	return s.Engine.Run(s.Config.ServerPort)
}

func SetupRouter(sh *controllers.StarHandler, ah *controllers.AuthHandler, seth *controllers.SettingsHandler) *gin.Engine {
	r := gin.Default()
	
	// 添加CORS中间件
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))
	
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	// 登录相关路由
	r.GET("/login", ah.LoginPage)
	r.GET("/logout", ah.Logout)
	r.POST("/auth/token-login", ah.TokenLogin)
	r.GET("/auth/github", ah.GitHubLogin)
	r.GET("/auth/github/callback", ah.GitHubCallback)

	// 需要认证的路由组
	auth := r.Group("/")
	auth.Use(ah.AuthMiddleware())
	{
		auth.GET("/", sh.IndexPage)
		auth.GET("/settings", seth.SettingsPage)

		api := auth.Group("/api")
		{
			api.GET("/user", sh.GetUser)
			api.GET("/repos", sh.GetRepos)
			api.GET("/stats", sh.GetStats)
			api.GET("/categories", sh.GetCategories)
			api.GET("/sync-progress", sh.SyncProgressWS)
			api.POST("/sync", sh.SyncStars)
			api.POST("/repos/:id/tag", sh.UpdateTag)
			api.POST("/repos/:id/category", sh.UpdateCategory)
			api.POST("/repos/:id/description", sh.UpdateDescription)
			api.POST("/repos/:id/analyze", sh.AnalyzeRepo)
			api.POST("/test-openai", seth.TestOpenAI)
			api.POST("/test-webdav", seth.TestWebDAV)
			api.GET("/settings", seth.GetSettings)
			api.POST("/settings", seth.SaveSettings)
		}

	}
	return r
}