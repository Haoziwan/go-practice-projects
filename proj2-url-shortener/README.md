# URL Shortener

一个使用 Gin 框架和 Redis 构建的 URL 短链接服务。

## 技术栈

- **Web 框架**: Gin
- **存储**: Redis
- **环境变量管理**: godotenv


## 功能特性

- ✅ 生成短链接：将长 URL 转换为短链接
- ✅ 重定向：通过短链接跳转到原始 URL
- ✅ 可配置过期时间：通过环境变量设置
- ✅ Redis 持久化存储
- ✅ 环境变量配置管理

## 环境变量配置

项目使用 `.env` 文件管理配置。首次使用时，复制 `.env.example` 并根据需要修改


## 快速开始

### 1. 使用 Docker 启动 Redis

```bash
docker-compose up -d
```

### 2. 配置环境变量

```bash
# 复制示例配置
cp .env.example .env

# 根据需要编辑 .env 文件
```

### 3. 安装依赖并运行

```bash
# 安装依赖
go mod tidy

# 运行应用
go run .
```

## API 接口

- `GET /` - 显示首页表单
- `POST /shorten` - 提交 URL 进行缩短
- `GET /:shortKey` - 短链接重定向到原始 URL
- `GET /static/*` - 静态文件服务



