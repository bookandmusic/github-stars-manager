# 🐳 Docker 部署指南

GitHub Stars Manager 支持通过 Docker 进行部署，提供了更简单的部署和更好的环境隔离。

## 自行构建镜像

使用项目中的 Dockerfile 构建镜像：

```bash
docker build -t github-stars-manager .
```

## 运行容器

### 基本运行

```bash
docker run -d \
  --name github-stars-manager \
  -p 8181:8181 \
  -e GITHUB_CLIENT_ID=your_client_id \
  -e GITHUB_CLIENT_SECRET=your_client_secret \
  -e GITHUB_REDIRECT_URL=http://localhost:8181/auth/github/callback \
  github-stars-manager:latest
```

### 持久化数据

为了保存数据，你需要挂载数据卷：

```bash
docker run -d \
  --name github-stars-manager \
  -p 8181:8181 \
  -e GITHUB_CLIENT_ID=your_client_id \
  -e GITHUB_CLIENT_SECRET=your_client_secret \
  -e GITHUB_REDIRECT_URL=http://localhost:8181/auth/github/callback \
  -v $(pwd)/data:/root/data \
  github-stars-manager:latest
```

## Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'
services:
  github-stars-manager:
    build: .
    container_name: github-stars-manager
    ports:
      - "8181:8181"
    environment:
      - GITHUB_CLIENT_ID=your_client_id
      - GITHUB_CLIENT_SECRET=your_client_secret
      - GITHUB_REDIRECT_URL=http://localhost:8181/auth/github/callback
    volumes:
      - ./data:/root/data
    restart: unless-stopped
```

启动服务：

```bash
docker-compose up -d
```
