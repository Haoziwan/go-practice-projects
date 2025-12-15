# 并发下载器 (Concurrent Downloader)

一个使用 Go 语言开发的高性能并发文件下载器，支持多线程下载大文件。

## 功能特性

- ✅ **并发下载**: 使用多个 goroutine 同时下载文件的不同部分
- ✅ **实时进度**: 显示下载进度、速度和已下载大小
- ✅ **分块下载**: 将文件分成多个块，提高下载速度
- ✅ **自动合并**: 下载完成后自动合并所有分块
- ✅ **错误处理**: 完善的错误处理和恢复机制
- ✅ **Range 请求支持**: 自动检测服务器是否支持 Range 请求

## 技术实现

### 核心原理

1. **分块策略**: 将文件分成 N 个块，每个 goroutine 负责下载一个块
2. **Range 请求**: 使用 HTTP Range 头请求文件的特定字节范围
3. **并发控制**: 使用 `sync.WaitGroup` 管理 goroutine 的生命周期
4. **进度追踪**: 使用 `sync.Mutex` 保护进度数据，实时显示下载状态
5. **文件合并**: 按顺序合并所有下载的块到最终文件
6. **错误处理**: 处理下载过程中的错误，如文件不存在、网络错误等


## 使用方法

### 基本用法

```bash
# 下载文件（使用默认 4 个线程）
go run . -url https://example.com/file.zip

# 指定输出文件名
go run . -url https://example.com/file.zip -output myfile.zip

# 指定并发线程数
go run . -url https://example.com/file.zip -workers 8
```

### 编译并运行

```bash
# 编译
go build -o downloader.exe

# 运行
./downloader.exe -url https://example.com/file.zip -workers 8
```

### 命令行参数

- `-url`: 要下载的文件 URL（必需）
- `-output`: 输出文件路径（可选，默认使用 URL 中的文件名）
- `-workers`: 并发下载的线程数（可选，默认为 4）

## 使用示例

```bash
# 示例 1: 下载一个大文件
go run . -url https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso -workers 8

# 示例 2: 下载并指定保存路径
go run . -url https://example.com/video.mp4 -output ./downloads/myvideo.mp4 -workers 6

# 示例 3: 使用单线程下载（适用于不支持 Range 的服务器）
go run . -url https://example.com/file.txt -workers 1
```

## 工作流程

1. **初始化**: 发送 HEAD 请求获取文件大小和服务器能力
2. **分块**: 根据文件大小和线程数计算每个块的范围
3. **并发下载**: 启动多个 goroutine，每个下载一个块
4. **进度监控**: 实时显示下载进度和速度
5. **合并文件**: 所有块下载完成后，按顺序合并到最终文件
6. **清理**: 删除临时文件和目录

