# 并发下载器

使用 Go + Wails 开发的高性能多线程文件下载工具。

## 功能特性

- ✅ 多线程并发下载，提升下载速度
- ✅ 实时显示下载进度和速度
- ✅ 简洁美观的图形界面
- ✅ 自动检测服务器 Range 支持

## 快速开始

### 运行开发版本

```bash
wails dev
```

### 构建应用

```bash
wails build
```

## 测试下载链接

```bash
# 100MB 测试文件
http://ipv4.download.thinkbroadband.com/100MB.zip

# 512MB 测试文件
http://ipv4.download.thinkbroadband.com/512MB.zip

# 1GB 测试文件
http://ipv4.download.thinkbroadband.com/1GB.zip
```

## 技术栈

- **后端**: Go - 并发下载核心逻辑
- **前端**: HTML + CSS + JavaScript - 简洁现代的 UI
- **框架**: Wails v2 - Go 桌面应用框架
- **并发**: Goroutines + Channels - 高效的并发控制

## 核心原理

1. 将文件分成多个块
2. 使用 HTTP Range 请求并发下载各个块
3. 实时追踪每个块的下载进度
4. 下载完成后自动合并所有块
