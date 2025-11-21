# ⚙️ 配置说明

GitHub Stars Manager 需要一些环境变量来进行配置。这些配置项可以通过系统环境变量或者在运行时通过命令行参数指定。

## 环境变量

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `GITHUB_CLIENT_ID` | 是 | 无 | GitHub OAuth App 的 Client ID |
| `GITHUB_CLIENT_SECRET` | 是 | 无 | GitHub OAuth App 的 Client Secret |
| `GITHUB_REDIRECT_URL` | 是 | http://localhost:8181/auth/github/callback | GitHub OAuth 回调地址 |
| `SERVER_PORT` | 否 | :8181 | 服务器监听端口 |
| `LOGGER_LEVEL` | 否 | info | 日志级别 (debug/info/warn/error) |

## 获取 GitHub OAuth 凭据

要使用 GitHub Stars Manager，你需要创建一个 GitHub OAuth App：

1. 访问 GitHub Settings → Developer settings → OAuth Apps
2. 点击 "New OAuth App"
3. 填写应用信息：
   - Application name: GitHub Stars Manager (或任意你喜欢的名字)
   - Homepage URL: http://localhost:8181 (或你的部署地址)
   - Authorization callback URL: http://localhost:8181/auth/github/callback (或你的部署地址对应的回调URL)
4. 点击 "Register application"
5. 记录下生成的 `Client ID` 和 `Client Secret`

## 配置 OpenAI (可选)

如果你想要使用 AI 分析功能，你需要一个 OpenAI API 密钥：

1. 访问 https://platform.openai.com/api-keys
2. 创建一个新的密钥
3. 在网站的配置页面，配置自己的信息