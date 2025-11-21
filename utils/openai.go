package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// Message OpenAI消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest OpenAI聊天请求结构
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse OpenAI聊天响应结构
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type OpenAIUtil struct {
	logger *zap.Logger
}

func NewOpenAIUtil(logger *zap.Logger) *OpenAIUtil {
	return &OpenAIUtil{logger: logger}
}

// mergeNestedField 将 "a.b.c" = value 合并进 map，不破坏结构
func (c *OpenAIUtil) mergeNestedField(data map[string]interface{}, key string, value interface{}) {
	keys := strings.Split(key, ".")
	current := data

	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = value
			return
		}

		next, exists := current[k]
		if !exists {
			newMap := make(map[string]interface{})
			current[k] = newMap
			current = newMap
			continue
		}

		m, ok := next.(map[string]interface{})
		if !ok {
			newMap := make(map[string]interface{})
			current[k] = newMap
			current = newMap
			continue
		}

		current = m
	}
}

func (o *OpenAIUtil) CallWithPrompt(settings OpenAISettings, prompt string) (string, error) {
	// 1. 序列化基础 body
	baseBody := ChatRequest{
		Model: settings.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}
	url := settings.Endpoint
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "chat/completions"
	bodyMap := make(map[string]interface{})
	tmpBytes, _ := json.Marshal(baseBody)
	json.Unmarshal(tmpBytes, &bodyMap)

	// 2. 合并自定义 body 字段
	for _, item := range settings.Body {
		o.mergeNestedField(bodyMap, item.Key, item.Value)
	}

	// 3. 序列化最终 body
	finalBody, err := json.Marshal(bodyMap)
	if err != nil {
		return "", fmt.Errorf("序列化失败: %w", err)
	}

	// 4. 构建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(finalBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 5. 设置 headers
	req.Header.Set("Content-Type", "application/json")
	for _, item := range settings.Headers {
		req.Header.Set(item.Key, item.Value)
	}

	// 6. 打印日志
	if o.logger != nil {
		o.logger.Debug("调用 Chat API",
			zap.String("url", url),
			zap.Any("headers", req.Header),
			zap.String("body", string(finalBody)),
		)
	}

	// 7. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	o.logger.Debug("OpenAI API响应状态", zap.Int("status", resp.StatusCode))

	// 错误状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("解析OpenAI响应失败: %w", err)
	}

	if chatResp.Error.Message != "" {
		return "", fmt.Errorf("OpenAI API返回错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI未返回有效结果")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (o *OpenAIUtil) TestConnection(settings OpenAISettings) error {
	prompt := "你好，请简单介绍一下你自己。"
	_, err := o.CallWithPrompt(settings, prompt)
	return err
}
