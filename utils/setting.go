package utils

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.yaml.in/yaml/v3"
)

// OpenAISettings OpenAI配置结构
type OpenAISettings struct {
	Key      string     `json:"key" yaml:"key"`
	Endpoint string     `json:"endpoint" yaml:"endpoint"`
	Model    string     `json:"model" yaml:"model"`
	Headers  []KeyValue `json:"headers" yaml:"headers"`
	Body     []KeyValue `json:"body" yaml:"body"`
}

// KeyValue 键值对结构，用于自定义请求头和请求体
type KeyValue struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

// WebDAVSettings WebDAV配置结构
type WebDAVSettings struct {
	Url      string `json:"url" yaml:"url"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// Settings 保存到文件的设置结构
type Settings struct {
	OpenAI OpenAISettings `json:"openai" yaml:"openai"`
	WebDAV WebDAVSettings `json:"webdav" yaml:"webdav"`
}

type SettingsUtil struct {
	logger *zap.Logger
}

func NewSettingsUtil(logger *zap.Logger) *SettingsUtil {
	return &SettingsUtil{
		logger: logger,
	}
}

// loadSettings 从文件加载设置
func (utl *SettingsUtil) LoadSettings() (*Settings, error) {
	settingsPath := filepath.Join("data", "settings.yaml")

	// 如果设置文件不存在，返回默认设置
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		utl.logger.Warn("配置文件不存在，使用默认设置")
		return &Settings{
			OpenAI: OpenAISettings{
				Key:      "",
				Endpoint: "",
				Model:    "",
				Headers:  make([]KeyValue, 0),
				Body:     make([]KeyValue, 0),
			},
			WebDAV: WebDAVSettings{
				Url:      "",
				Username: "",
				Password: "",
			},
		}, nil
	}

	// 读取设置文件
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		utl.logger.Error("读取配置文件失败", zap.Error(err))
		return nil, err
	}
	utl.logger.Info("读取配置文件成功")
	// 解析YAML
	var settings Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		utl.logger.Error("解析配置文件失败", zap.Error(err))
		return nil, err
	}

	// 确保Headers和Body不为nil
	if settings.OpenAI.Headers == nil {
		settings.OpenAI.Headers = make([]KeyValue, 0)
	}
	if settings.OpenAI.Body == nil {
		settings.OpenAI.Body = make([]KeyValue, 0)
	}
	utl.logger.Info("解析配置文件成功")
	return &settings, nil
}

func (utl *SettingsUtil) SaveSettings(settings *Settings) error {
	// 判断data目录是否存在，不存在，创建
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		utl.logger.Info("创建data目录")
		if err := os.Mkdir("data", 0755); err != nil {
			utl.logger.Error("创建data目录失败", zap.Error(err))
			return err
		}
	}
	settingsPath := filepath.Join("data", "settings.yaml")

	// 将设置转换为YAML
	data, err := yaml.Marshal(settings)
	if err != nil {
		utl.logger.Error("将设置转换为YAML失败", zap.Error(err))
		return err
	}

	// 创建或覆盖设置文件
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		utl.logger.Error("创建或覆盖设置文件失败", zap.Error(err))
		return err
	}
	utl.logger.Info("保存设置成功")
	return nil
}
