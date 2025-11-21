package di

import (
	"github-stars-manager/config"
	"github-stars-manager/controllers"
	"github-stars-manager/logger"
	"github-stars-manager/repository"
	"github-stars-manager/routes"
	"github-stars-manager/utils"

	"go.uber.org/dig"
)

// Container 依赖注入容器
func NewContainer() *dig.Container {
	Container := dig.New()

	// 提供配置
	Container.Provide(config.NewConfig)

	// 提供日志记录器
	Container.Provide(logger.NewDevelopmentLogger)

	Container.Provide(utils.NewOpenAIUtil)

	Container.Provide(utils.NewSettingsUtil)

	Container.Provide(utils.NewGithubCli)

	// 提供数据仓库
	Container.Provide(repository.NewFileRepository)

	// 提供StarHandler
	Container.Provide(controllers.NewStarHandler)

	// 提供AuthHandler
	Container.Provide(controllers.NewAuthHandler)

	Container.Provide(controllers.NewSettingsHandler)

	// 提供路由引擎
	Container.Provide(routes.SetupRouter)

	Container.Provide(routes.NewServer)
	return Container
}

