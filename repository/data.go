package repository

import (
	"github-stars-manager/utils"
)

// RepoTag 代表仓库的标签和分类信息
type RepoTag struct {
	ID          int64  `json:"id"`
	Tag         string `json:"tag"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
}

// Repository 定义数据访问接口
type Repository interface {
	// GetReposWithTag 获取带标签信息的仓库列表
	GetReposWithTag() ([]utils.Repo, error)
	
	// SaveRepos 保存仓库列表
	SaveRepos(repos []utils.Repo) error
	
	// GetRepoTag 获取特定仓库的标签信息
	GetRepoTag(repoID int64) (*RepoTag, error)
	
	// SaveRepoTag 保存仓库标签信息
	SaveRepoTag(tag *RepoTag) error
	
	// DeleteRepoTag 删除仓库标签信息
	DeleteRepoTag(repoID int64) error
	
	// GetStats 获取统计信息
	GetStats() (*Stats, error)
	
	// SaveSyncTime 保存同步时间
	SaveSyncTime() error
	
	// LoadSyncTime 加载同步时间
	LoadSyncTime() (string, error)
}

// Stats 统计信息
type Stats struct {
	TotalRepos    int    `json:"total_repos"`
	AnalyzedRepos int    `json:"analyzed_repos"`
	LastSync      string `json:"last_sync"`
}