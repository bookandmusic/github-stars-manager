package main

import (
	"github-stars-manager/di"
	"github-stars-manager/routes"
)

func main() {
	// 从DIG容器中获取服务器实例并启动服务
	container := di.NewContainer()
	err := container.Invoke(func(server *routes.Server) error {
		return server.Run()
	})
	if err != nil {
		panic(err)
	}
}
