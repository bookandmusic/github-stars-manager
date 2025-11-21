# ⭐ GitHub Stars Manager

GitHub Stars Manager 可以帮助用户更好地分类、搜索与管理在 GitHub 上收藏的仓库。
主要特点包括同步 star 列表、手工/自动标签、AI 分析建议和本地数据存储以保护隐私。

---

## 📸 截图

### 登录界面
![Login](docs/screenshots/login.png)

### 仪表板界面
![Dashboard](docs/screenshots/dashboard.png)

### 设置界面
![Settings](docs/screenshots/settings.png)

---

## 🚀 快速开始

### 使用 Docker

```bash
docker run -d \
  --name github-stars-manager \
  -p 8181:8181 \
  -e GITHUB_CLIENT_ID=your_client_id \
  -e GITHUB_CLIENT_SECRET=your_client_secret \
  -e GITHUB_REDIRECT_URL=http://localhost:8181/auth/github/callback \
  -v $(pwd)/data:/root/data \
  ghcr.io/bookandmusic/github-stars-manager:latest
```

打开浏览器访问 `http://localhost:8181`

### 使用 Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'
services:
  github-stars-manager:
    image: ghcr.io/bookandmusic/github-stars-manager:latest
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

---

## 📚 文档

- [配置说明](docs/config.md) - 环境变量和配置项说明
- [本地开发](docs/build.md) - 如何在本地构建和运行
- [Docker 部署](docs/docker.md) - 使用 Docker 部署的详细指南

---

## 📝 许可证

MIT

---

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=bookandmusic/github-stars-manager&type=date&legend=top-left)](https://www.star-history.com/#bookandmusic/github-stars-manager&type=date&legend=top-left)

---

## 📣 贡献

欢迎提交 Issue、PR 或讨论新的功能。请在 PR 中包含可复现的测试步骤与简要说明。