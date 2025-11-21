package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	RedirectURL        string
	ServerPort         string
	LoggerLevel        string
}

// NewConfig 从环境变量创建配置实例
func NewConfig() *Config {
	// 设置默认值
	viper.SetDefault("LOGGER_LEVEL", "info")
	viper.SetDefault("GITHUB_CLIENT_ID", "你的ClientID")
	viper.SetDefault("GITHUB_CLIENT_SECRET", "你的ClientSecret")
	viper.SetDefault("GITHUB_REDIRECT_URL", "http://localhost:8181/auth/github/callback")
	viper.SetDefault("SERVER_PORT", ":8181")

	// 从环境变量中读取配置
	viper.AutomaticEnv()

	// 注意：launch.json中的环境变量名称需要与viper配置的名称一致
	return &Config{
		GitHubClientID:     viper.GetString("GITHUB_CLIENT_ID"),
		GitHubClientSecret: viper.GetString("GITHUB_CLIENT_SECRET"),
		RedirectURL:        viper.GetString("GITHUB_REDIRECT_URL"),
		ServerPort:         viper.GetString("SERVER_PORT"),
		LoggerLevel:        viper.GetString("LOGGER_LEVEL"),
	}
}