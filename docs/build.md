# 🛠 本地开发指南

本文档介绍了如何在本地环境中设置和运行 GitHub Stars Manager。

## 环境要求

- Go 1.25+
- Git

## 克隆项目

```bash
git clone https://github.com/bookandmusic/github-stars-manager.git
cd github-stars-manager
```

## 配置环境变量

参考 [配置说明](config.md) 设置必要的环境变量。

## 运行项目

开发模式下可以直接运行：

```bash
go run main.go
```

或者构建后再运行：

```bash
go build -o github-stars-manager
./github-stars-manager
```

## 构建参数

项目支持通过构建参数来自定义构建：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o github-stars-manager
```

## 依赖管理

项目使用 Go Modules 进行依赖管理。如果需要添加新依赖：

```bash
go get package-name
```

更新依赖：

```bash
go mod tidy
```