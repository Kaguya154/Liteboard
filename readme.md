### 使用方法

#### 配置环境

安装 go 1.25+

配置环境
```bash
go mod tidy
```

#### 运行

生成 Swagger 文档

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency --parseInternal --parseDepth 5 --instanceName "swagger"
```

运行项目

```bash
go run main.go
```

#### 生成可执行文件

```bash
go build -o app main.go
```

