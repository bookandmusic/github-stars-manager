package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"go.uber.org/zap"
)

type GithubUtil struct {
	logger *zap.Logger
}

func NewGithubCli(logger *zap.Logger) *GithubUtil {
	return &GithubUtil{
		logger: logger,
	}
}

type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type Repo struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	HTMLURL         string   `json:"html_url"`
	StargazersCount int      `json:"stargazers_count"`
	Description     string   `json:"description"`
	Language        string   `json:"language"`
	Languages       []string `json:"languages"`
	Topics          []string `json:"topics"`
	Tag             string   `json:"tag"`
	Category        string   `json:"category"`
	ReadmeURL       string   `json:"readme_url"`
}

// GetAccessToken 获取GitHub access token
func (utl *GithubUtil)GetAccessToken(clientID, clientSecret, code string) (string, error) {
	utl.logger.Debug("获取GitHub access token")
	url := "https://github.com/login/oauth/access_token"
	payload := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		utl.logger.Error("获取access token请求失败", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		utl.logger.Error("解析access token响应失败", zap.Error(err))
		return "", err
	}

	utl.logger.Debug("获取GitHub access token成功")
	return result.AccessToken, nil
}

// GetUserInfo 获取用户信息
func (utl *GithubUtil) GetUserInfo(token string) (*User, error) {
	utl.logger.Debug("获取GitHub用户信息")
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		utl.logger.Error("获取用户信息请求失败", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		utl.logger.Error("解析用户信息响应失败", zap.Error(err))
		return nil, err
	}

	utl.logger.Debug("获取GitHub用户信息成功", zap.String("user", user.Login))
	return &user, nil
}

// GetStarredRepos 获取用户star的仓库列表
func  (utl *GithubUtil)GetStarredRepos(token string) ([]Repo, error) {
	utl.logger.Debug("获取用户star的仓库列表")
	var allRepos []Repo
	page := 1

	// 循环获取所有页面的仓库
	for {
		client := &http.Client{Timeout: 30 * time.Second}
		url := fmt.Sprintf("https://api.github.com/user/starred?page=%d&per_page=100", page)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			utl.logger.Error("获取star仓库列表失败", 
				zap.Error(err), 
				zap.Int("page", page))
			return nil, err
		}

		var repos []Repo
		err = json.NewDecoder(resp.Body).Decode(&repos)
		resp.Body.Close()
		if err != nil {
			utl.logger.Error("解析star仓库列表失败", 
				zap.Error(err), 
				zap.Int("page", page))
			return nil, err
		}

		// 添加到总列表中
		allRepos = append(allRepos, repos...)

		// 如果当前页面的仓库数量少于100，说明已经是最后一页
		if len(repos) < 100 {
			break
		}

		page++
	}

	utl.logger.Debug("获取用户star的仓库列表成功", zap.Int("count", len(allRepos)))
	return allRepos, nil
}

// GetRepoDetails 获取仓库详细信息
func  (utl *GithubUtil)GetRepoDetails(token, repoFullName string) (*Repo, error) {
	utl.logger.Debug("获取仓库详细信息", zap.String("repo", repoFullName))
	client := &http.Client{Timeout: 30 * time.Second}
	
	// 获取仓库基本信息
	url := fmt.Sprintf("https://api.github.com/repos/%s", repoFullName)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		utl.logger.Error("获取仓库信息失败", 
			zap.Error(err), 
			zap.String("repo", repoFullName))
		return nil, err
	}

	var repo Repo
	err = json.NewDecoder(resp.Body).Decode(&repo)
	resp.Body.Close()
	if err != nil {
		utl.logger.Error("解析仓库信息失败", 
			zap.Error(err), 
			zap.String("repo", repoFullName))
		return nil, err
	}

	// 获取仓库语言信息
	langURL := fmt.Sprintf("https://api.github.com/repos/%s/languages", repoFullName)
	langReq, _ := http.NewRequest("GET", langURL, nil)
	langReq.Header.Set("Authorization", "token "+token)
	langReq.Header.Set("Accept", "application/vnd.github.v3+json")

	langResp, err := client.Do(langReq)
	if err != nil {
		utl.logger.Warn("获取仓库语言信息失败", 
			zap.Error(err), 
			zap.String("repo", repoFullName))
	} else {
		var languages map[string]int
		err = json.NewDecoder(langResp.Body).Decode(&languages)
		langResp.Body.Close()
		if err != nil {
			utl.logger.Warn("解析仓库语言信息失败", 
				zap.Error(err), 
				zap.String("repo", repoFullName))
		} else {
			// 将语言map转换为语言列表
			for lang := range languages {
				repo.Languages = append(repo.Languages, lang)
			}
		}
	}
	
	// 构造README链接
	repo.ReadmeURL = fmt.Sprintf("%s#readme", repo.HTMLURL)

	utl.logger.Debug("获取仓库详细信息成功", zap.String("repo", repoFullName))
	return &repo, nil
}
