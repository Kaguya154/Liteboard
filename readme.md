# Liteboard

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Kaguya154/liteboard)](https://goreportcard.com/report/github.com/Kaguya154/liteboard)

轻量级内容看板与权限管理服务。后端使用 CloudWeGo Hertz，内置 SQLite 存储，支持 GitHub OAuth 登录、会话权限校验，并提供完整的 RESTful API 与 Swagger 文档，以及简洁的前端页面（首页、仪表盘、看板与分享页）。

- 语言与运行时：Go 1.25+
- Web 框架：CloudWeGo Hertz
- 存储：SQLite（本地文件 liteboard.db）
- 文档：Swagger UI（/swagger）

## 功能特性

- GitHub OAuth 登录，基于 Cookie Session 的会话管理
- 用户与分组（示例组：user/admin），中间件进行权限校验
- 内容模块：Content List、Content Entry 的增删改查
- 项目、权限、分享 Token 等 API 能力（详见 Swagger）
- 自动建表，零额外数据库安装，启动即用
- 自带静态页面与简单看板（/、/dashboard、/board.html、/share）

## 快速开始

### 环境要求

- 安装 Go 1.25+（以 go.mod 为准）

### 克隆与依赖

- 将本仓库克隆到本地后，在项目根目录执行：

```cmd
go mod tidy
```

### 配置（必读）

本项目通过 .env 读取 GitHub OAuth 配置（auth/openid.go 使用了 github.com/joho/godotenv）。示例：

```env
# .env（不要提交到仓库）
GITHUB_CLIENT_ID=你的GitHub客户端ID
GITHUB_CLIENT_SECRET=你的GitHub客户端密钥
# 可选，默认为 http://localhost:8080/auth/github/callback
GITHUB_REDIRECT_URI=http://localhost:8080/auth/github/callback
```

注意：不要将真实密钥提交到版本库。生产环境建议通过环境变量注入，不使用 .env 文件。

### 生成 Swagger（如需更新文档）

项目已内置 docs，如你修改了接口注释需要重新生成：

```cmd
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency --parseInternal --parseDepth 5 --instanceName "swagger"
```

完成后可通过 Swagger UI 访问 API 文档：

- http://localhost:8080/swagger/index.html

提示：运行服务时需携带 `-swagger` 参数才会开启 /swagger 路由。

### 运行

默认监听 0.0.0.0:8080，可通过命令行参数修改：

- -a 监听地址（默认 0.0.0.0）
- -p 端口（默认 8080）
- -s 会话密钥（默认 secret，强烈建议生产环境指定强随机值）
- -swagger 启用 Swagger 文档（默认关闭）

启动服务（开发示例，启用 Swagger，并设置会话密钥）：

```cmd
go run main.go -s "change-me-please" -swagger
```

或指定端口/地址：

```cmd
go run main.go -a 127.0.0.1 -p 9000 -s "change-me-please" -swagger
```

启动后：

- 访问首页：http://localhost:8080/
- 登录（GitHub OAuth）：访问 /auth/github/login 或首页跳转
- 仪表盘（需登录）：http://localhost:8080/dashboard
- Swagger 文档（启用 -swagger 时）：http://localhost:8080/swagger/index.html

### TLS 配置（可选，高级）

使用 `-tls` 可启用 TLS；需要提供：

- `-crt` 服务器证书路径（默认 server.crt）
- `-key` 服务器私钥路径（默认 server.key）
- `-ca` 客户端证书签发的 CA 证书路径（默认 ca.crt）

注意：当前实现启用了“要求并验证客户端证书”（mTLS）。启用 `-tls` 后，客户端必须携带由 `-ca` 指定根证书签发的客户端证书才能成功访问。

示例（在 8443 端口启用 TLS 与 Swagger）：

```cmd
go run main.go -p 8443 -tls -crt server.crt -key server.key -ca ca.crt -s "change-me-please" -swagger
```

生产环境建议使用受信任 CA 颁发的证书，并妥善管理私钥与 CA。

### 构建可执行文件

```cmd
REM Windows
go build -o liteboard.exe main.go

REM Linux / macOS（如在对应平台构建）
go build -o liteboard main.go
```

## 命令行参数

- `-a` 监听地址，默认 `0.0.0.0`
- `-p` 监听端口，默认 `8080`
- `-h` 显示帮助
- `-s` 会话密钥（Cookie Session 加密/签名），默认 `secret`。生产务必设置为强随机值
- `-swagger` 是否启用 Swagger 文档路由，默认关闭
- `-tls` 是否启用 TLS（启用后默认要求并验证客户端证书）
- `-crt` TLS 服务器证书路径，默认 `server.crt`
- `-key` TLS 服务器私钥路径，默认 `server.key`
- `-ca` TLS CA 证书路径（用于验证客户端证书），默认 `ca.crt`

## 配置说明

- 数据库：使用 SQLite，数据文件默认为项目根目录下的 liteboard.db；首次启动会自动建表
- 会话：Cookie Session，使用 `-s` 指定密钥；默认 `secret` 仅用于开发，生产务必更换为强随机值
- 管理员示例：代码中包含授予特定用户 admin 组的示例逻辑（auth/openid.go）。上线前请按需修改/移除

## API 概览（节选）

更多请以 Swagger 为准，仅列出关键接口示例：

- 用户
  - GET /api/user/profile 获取当前登录用户信息

- 内容列表（Content List）
  - GET /api/content_lists?projectid={id}
  - POST /api/content_lists
  - GET /api/content_lists/{id}
  - PUT /api/content_lists/{id}
  - DELETE /api/content_lists/{id}

- 内容条目（Content Entry）
  - GET /api/content_entries
  - POST /api/content_entries
  - GET /api/content_entries/{id}
  - PUT /api/content_entries/{id}
  - DELETE /api/content_entries/{id}

- 其他：项目、权限、分享 Token 等接口已注册，可在 Swagger 中查看

权限与认证：

- /api 路由受登录与分组权限保护（示例需要具备 user 或 admin），部分接口还会做细粒度内容权限校验

## 测试

运行全部测试：

```cmd
go test ./...
```

## 目录结构（简要）

- frontend/ 静态页面与资源
- api/ 业务路由与处理器
- auth/ 登录、会话与权限相关逻辑（GitHub OAuth）
- internal/ 数据模型、CRUD、权限校验等
- docs/ Swagger 生成产物

## 贡献指南

欢迎提交 Issue 与 PR：

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/your-feature`
3. 提交变更：`git commit -m "feat: your message"`
4. 推送分支：`git push origin feature/your-feature`
5. 发起 Pull Request

## 许可协议

本项目基于 MIT 许可证发布，详见 LICENSE。

## 致谢与安全提示

- 本项目基于 CloudWeGo Hertz 与 Swag 等开源组件，感谢社区
- 请勿将任何密钥（如 .env）提交至仓库；生产环境请更换会话密钥并完善管理员判定逻辑
