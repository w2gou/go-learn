# Go-Transfer

🚀 **Go-Transfer: 二维码驱动的局域网临时文件分享CLI工具**

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Platform](https://img.shields.io/badge/Platform-Windows%7CLinux%7CmacOS-brightgreen.svg)

## ✨ 核心特性

### 🚀 高性能并发处理
- **Goroutine驱动**: 每个请求独立的Goroutine处理，支持高并发
- **连接池管理**: 智能连接数限制，防止资源耗尽
- **性能监控**: 实时统计连接数、传输量、吞吐量

### 📱 智能二维码生成
- **多尺寸支持**: small/medium/large 三种二维码大小
- **ASCII显示**: 终端原生显示，无需外部工具
- **自动网络发现**: 智能获取局域网IP地址

### ⚙️ 专业CLI体验
- **Cobra框架**: 现代化的命令行接口
- **Viper配置**: 支持配置文件和环境变量
- **优雅关闭**: 信号处理和资源清理

### �️ 企业级安全
- **HTTP基础认证**: 可选的用户名/密码保护
- **访问控制**: IP限制和连接数控制
- **安全传输**: 支持HTTPS（可配置）

### 📊 可观测性
- **结构化日志**: 基于zap的高性能日志系统
- **性能指标**: CPU、内存、网络统计
- **实时监控**: /api/stats 和 /api/health 端点

## 🏗️ 技术架构

### 核心技术栈
- **语言**: Go 1.21+
- **CLI框架**: Cobra + Viper
- **HTTP框架**: Gorilla Mux
- **日志系统**: Uber Zap
- **二维码**: go-qrcode

### 架构设计
```
┌─────────────────────────────────────────────────────┐
│                   Go-Transfer                       │
├─────────────────────────────────────────────────────┤
│  CLI Layer (Cobra + Viper)                         │
│  ├── serve      ├── config     ├── version         │
├─────────────────────────────────────────────────────┤
│  Business Layer                                     │
│  ├── Server     ├── QR Code    ├── Network         │
│  ├── Config     ├── Auth       ├── Stats           │
├─────────────────────────────────────────────────────┤
│  Infrastructure Layer                               │
│  ├── HTTP Server (Gorilla Mux)                     │
│  ├── Concurrent Handler (Goroutine Pool)           │
│  ├── Logging (Zap)                                 │
│  ├── Monitoring (Metrics)                          │
└─────────────────────────────────────────────────────┘
```

### 并发模型
```go
// 每个HTTP请求的处理流程
Client Request → Semaphore → Goroutine Pool → Handler
                    ↓              ↓           ↓
                 限流控制      并发处理      业务逻辑
                    ↓              ↓           ↓
                 监控统计      日志记录      响应返回
```

## 🚀 快速开始

### 安装方式

#### 方式1: 源码编译
```bash
git clone <repository-url>
cd Go-Transfer
make build
```

#### 方式2: 使用Makefile（推荐）
```bash
# 构建开发版本
make build

# 构建生产版本（优化）
make build-prod

# 交叉编译所有平台
make dist
```

#### 方式3: Go安装
```bash
go install <repository-url>@latest
```

### 基本使用

#### 🎯 命令概览
```bash
# 查看帮助
go-transfer --help          # Linux/macOS
.\go-transfer.exe --help    # Windows PowerShell

# 分享单个文件
go-transfer serve document.pdf          # Linux/macOS
.\go-transfer.exe serve document.pdf    # Windows PowerShell

# 分享文件夹（默认当前目录）
go-transfer serve ./photos          # Linux/macOS  
.\go-transfer.exe serve ./photos    # Windows PowerShell

# 高级选项
go-transfer serve --port 9000 --auth --verbose ./files          # Linux/macOS
.\go-transfer.exe serve --port 9000 --auth --verbose ./files    # Windows PowerShell
```

> **Windows用户注意**: PowerShell出于安全考虑，不会自动从当前目录执行程序。必须使用 `.\go-transfer.exe` 格式，或将程序安装到系统PATH中。详见 [Windows使用指南](docs/WINDOWS_USAGE.md)。

#### 🔧 配置选项

**命令行参数**:
```bash
--port        服务器端口 (默认: 8080)
--auth        启用HTTP基础认证
--username    认证用户名 (默认: admin)
--password    认证密码 (空则随机生成)
--verbose     详细日志输出
--no-qr       禁用二维码显示
--max-connections  最大并发连接数 (默认: 100)
```

**配置文件**: `~/.go-transfer.yaml`
```yaml
server:
  port: "8080"
  auth: false
  max_connections: 100
  timeout: "30s"

qr:
  disabled: false
  size: "medium"

logging:
  verbose: false
  format: "text"
```

## 📊 性能特性

### 并发能力
- **连接池**: 可配置最大并发连接数
- **Goroutine管理**: 高效的协程调度
- **内存优化**: 零拷贝文件传输
- **负载均衡**: 智能请求分发

### 性能基准
```bash
# 运行性能测试
make benchmark

# 示例结果 (仅供参考)
BenchmarkServer_HandleRequest-8      5000   250ns/op   0 allocs/op
BenchmarkServer_ConcurrentConnections-8  1000  1.2μs/op  48 B/op
```

### 监控端点
```bash
# 健康检查
curl http://localhost:8080/api/health

# 性能统计
curl http://localhost:8080/api/stats
```

## 🛠️ 开发指南

### 项目结构
```
Go-Transfer/
├── cmd/                    # CLI命令定义
│   ├── root.go            # 根命令和全局配置
│   └── serve.go           # serve子命令
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── network/          # 网络工具
│   ├── qr/              # 二维码生成
│   └── server/           # HTTP服务器
├── pkg/                  # 公共包
├── tests/               # 集成测试
├── configs/             # 配置文件
├── Makefile            # 构建脚本
├── go.mod              # Go模块
└── README.md           # 项目文档
```

### 开发工作流
```bash
# 1. 安装依赖
make deps

# 2. 代码格式化
make fmt

# 3. 静态检查
make vet lint

# 4. 运行测试
make test

# 5. 性能测试
make benchmark

# 6. 构建发布
make release
```

### 测试覆盖率
```bash
# 生成测试覆盖率报告
make coverage

# 查看HTML报告
open coverage/coverage.html
```

## 🔧 高级特性

### 1. 企业级日志系统
- **结构化日志**: JSON格式，便于日志分析
- **日志级别**: Debug/Info/Warn/Error
- **性能优化**: 零分配日志记录

### 2. 性能监控
- **实时统计**: 连接数、请求数、传输字节数
- **系统指标**: CPU使用率、内存占用、Goroutine数量
- **客户端分析**: IP统计、访问热点文件

### 3. 安全特性
- **认证系统**: 可选的HTTP基础认证
- **访问控制**: 基于IP的访问限制
- **安全传输**: 支持HTTPS配置
- **防攻击**: 请求频率限制

### 4. 可扩展性
- **插件系统**: 支持自定义中间件
- **配置热加载**: 运行时配置更新
- **多实例部署**: 负载均衡支持

## 🌍 部署方案

### 1. 单机部署
```bash
# 直接运行
./go-transfer serve --port 8080 ./files

# 后台运行
nohup ./go-transfer serve ./files > transfer.log 2>&1 &
```

### 2. Docker部署
```dockerfile
FROM alpine:latest
COPY go-transfer /usr/local/bin/
EXPOSE 8080
CMD ["go-transfer", "serve", "/data"]
```

### 3. 系统服务
```bash
# 创建systemd服务
sudo cp go-transfer.service /etc/systemd/system/
sudo systemctl enable go-transfer
sudo systemctl start go-transfer
```

## 🤝 贡献指南

### 开发环境
1. Go 1.21+
2. Make工具
3. Git版本控制

### 贡献流程
1. Fork项目
2. 创建功能分支
3. 提交代码更改
4. 添加测试用例
5. 提交Pull Request

### 代码规范
- 遵循Go官方编码规范
- 添加必要的注释和文档
- 确保测试覆盖率 > 80%
- 通过所有静态检查

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

## 🎉 致谢

感谢以下开源项目的贡献：
- [Cobra](https://github.com/spf13/cobra) - 现代化CLI框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP路由
- [Zap](https://github.com/uber-go/zap) - 高性能日志
- [go-qrcode](https://github.com/skip2/go-qrcode) - 二维码生成

---

**⭐ 如果这个项目对您有帮助，请给我们一个Star！**