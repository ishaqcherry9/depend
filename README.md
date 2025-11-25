# 目录

- [目录](#目录)
- [depend](#depend)
  - [📦 安装](#-安装)
  - [🚀 功能特性](#-功能特性)
    - [核心组件](#核心组件)
      - [缓存 (Cache)](#缓存-cache)
      - [日志 (Logger)](#日志-logger)
      - [对象存储 (OSS)](#对象存储-oss)
      - [数据库 (GORM)](#数据库-gorm)
      - [配置管理 (Config)](#配置管理-config)
      - [消息队列 (Pulsar)](#消息队列-pulsar)
    - [Web 框架支持](#web-框架支持)
      - [Gin 中间件](#gin-中间件)
    - [认证授权](#认证授权)
      - [JWT](#jwt)
      - [Token 管理 (CToken)](#token-管理-ctoken)
    - [安全加密](#安全加密)
      - [加密工具 (Gocrypto)](#加密工具-gocrypto)
    - [限流熔断](#限流熔断)
      - [保护机制 (Shield)](#保护机制-shield)
    - [工具类](#工具类)
      - [HTTP 客户端](#http-客户端)
      - [JSON 处理](#json-处理)
      - [对象拷贝 (Copier)](#对象拷贝-copier)
      - [定时任务 (Gocron)](#定时任务-gocron)
      - [验证码 (Captcha)](#验证码-captcha)
      - [工具函数 (Utils)](#工具函数-utils)
    - [分布式组件](#分布式组件)
      - [分布式锁 (DLock)](#分布式锁-dlock)
      - [负载均衡 (LBA)](#负载均衡-lba)
    - [可观测性](#可观测性)
      - [链路追踪 (Tracer)](#链路追踪-tracer)
      - [性能分析 (Prof)](#性能分析-prof)
      - [监控统计 (Stat)](#监控统计-stat)
    - [RPC 客户端](#rpc-客户端)
      - [IM-Proxy SDK](#im-proxy-sdk)
    - [其他组件](#其他组件)
  - [📚 使用示例](#-使用示例)
    - [基础使用](#基础使用)
    - [Gin 应用示例](#gin-应用示例)
  - [🔧 配置](#-配置)
  - [📝 更新日志](#-更新日志)
  - [🤝 贡献](#-贡献)
  - [📄 许可证](#-许可证)
  - [🔗 相关链接](#-相关链接)

# depend

一个功能丰富的 Go 语言基础工具库集合，提供了开发中常用的各种工具和组件。

## 📦 安装

```bash
go get github.com/ishaqcherry9/depend
```

## 🚀 功能特性

### 核心组件

#### 缓存 (Cache)
- **内存缓存**：基于 Ristretto 的高性能内存缓存
- **Redis 缓存**：支持单机、集群模式的 Redis 缓存
- **分层缓存**：内存 + Redis 的分层缓存方案
- **分布式锁**：基于 Redis 的分布式锁实现

```go
import "github.com/ishaqcherry9/depend/pkg/cache"

// 使用缓存
cache.Set(ctx, "key", "value", time.Hour)
cache.Get(ctx, "key", &value)
```

#### 日志 (Logger)
- 基于 Zap 的高性能日志库
- 支持控制台和文件输出
- 支持 JSON 和 Console 格式
- 自动日志轮转

```go
import "github.com/ishaqcherry9/depend/pkg/logger"

logger.Info("info message")
logger.Error("error message")
```

#### 对象存储 (OSS)
- **AWS S3**：完整的 S3 客户端支持
- **MinIO**：MinIO 对象存储支持
- **文件处理**：图片裁剪、缩略图生成、视频封面提取
- **分片上传**：支持大文件分片并发上传

```go
import "github.com/ishaqcherry9/depend/pkg/oss"

// 上传文件
resp, err := oss.Upload(ctx, req)
```

#### 数据库 (GORM)
- 基于 GORM 的数据库封装
- 支持 MySQL 连接池管理
- 自动重连和健康检查
- 支持读写分离

```go
import "github.com/ishaqcherry9/depend/pkg/sgorm"

// 使用数据库
db := sgorm.GetDB()
```

#### 配置管理 (Config)
- **Viper**：基于 Viper 的配置管理
- **Nacos**：Nacos 配置中心集成
- 支持多格式配置文件（YAML、JSON、TOML）
- 配置热更新

```go
import "github.com/ishaqcherry9/depend/pkg/conf"

// 读取配置
config := conf.GetConfig()
```

#### 消息队列 (Pulsar)
- Apache Pulsar 客户端封装
- 支持生产者和消费者
- 消息重试和死信队列

```go
import "github.com/ishaqcherry9/depend/pkg/pulsar"

// 创建生产者
producer, _ := pulsar.NewProducer(config)
```

### Web 框架支持

#### Gin 中间件
- **认证中间件**：JWT Token 验证
- **限流中间件**：基于令牌桶的限流
- **CORS**：跨域资源共享支持
- **Swagger**：自动生成 API 文档
- **性能分析**：pprof 性能分析集成
- **请求日志**：请求/响应日志记录

```go
import "github.com/ishaqcherry9/depend/pkg/gin/middleware"

// 使用中间件
router.Use(middleware.Auth())
router.Use(middleware.RateLimit())
```

### 认证授权

#### JWT
- JWT Token 生成和验证
- 支持多种签名算法
- Token 刷新机制

```go
import "github.com/ishaqcherry9/depend/pkg/jwt"

// 生成 Token
token, _ := jwt.GenerateToken(claims)
```

#### Token 管理 (CToken)
- Token 缓存管理
- Redis 适配器
- Token 编码/解码

### 安全加密

#### 加密工具 (Gocrypto)
- **AES**：AES 加密/解密
- **DES**：DES 加密/解密
- **RSA**：RSA 非对称加密
- **哈希**：MD5、SHA256 等哈希算法
- **密码**：密码加密和验证

```go
import "github.com/ishaqcherry9/depend/pkg/gocrypto"

// AES 加密
encrypted, _ := gocrypto.AESEncrypt(data, key)
```

### 限流熔断

#### 保护机制 (Shield)
- **限流**：令牌桶、滑动窗口限流
- **熔断器**：断路器模式实现
- **CPU 保护**：CPU 使用率监控和保护
- **窗口统计**：时间窗口统计

```go
import "github.com/ishaqcherry9/depend/pkg/shield/ratelimit"

// 创建限流器
limiter := ratelimit.NewLimiter(100, time.Second)
```

### 工具类

#### HTTP 客户端
- 统一的 HTTP 客户端封装
- 支持超时、重试
- 请求/响应拦截器

```go
import "github.com/ishaqcherry9/depend/pkg/httpcli"

// 发送请求
resp, err := httpcli.Get(ctx, url)
```

#### JSON 处理
- 高性能 JSON 序列化/反序列化
- 基于 json-iterator
- 类型安全的 JSON 操作

```go
import "github.com/ishaqcherry9/depend/pkg/jsonx"

// JSON 操作
jsonx.Marshal(data)
jsonx.Unmarshal(jsonData, &result)
```

#### 对象拷贝 (Copier)
- 结构体字段拷贝
- 支持嵌套结构
- 类型转换

#### 定时任务 (Gocron)
- 基于 cron 表达式的定时任务
- 支持秒级精度
- 任务日志记录

```go
import "github.com/ishaqcherry9/depend/pkg/gocron"

// 创建定时任务
cron := gocron.New()
cron.AddFunc("0 0 * * *", func() { /* ... */ })
```

#### 验证码 (Captcha)
- 图形验证码生成
- 易盾验证码集成
- Base64 编码支持

#### 工具函数 (Utils)
- 字符串处理
- 时间处理
- UUID 生成
- 随机数生成

### 分布式组件

#### 分布式锁 (DLock)
- 基于 Redis 的分布式锁
- 支持自动续期
- 锁超时机制

```go
import "github.com/ishaqcherry9/depend/pkg/dlock"

// 获取锁
lock := dlock.NewLock(key)
lock.Lock()
defer lock.Unlock()
```

#### 负载均衡 (LBA)
- **轮询**：Round Robin
- **加权轮询**：Weighted Round Robin
- **随机**：Random
- **最少连接**：Least Connections
- **IP 哈希**：IP Hash
- **P2C**：Power of Two Choices

### 可观测性

#### 链路追踪 (Tracer)
- OpenTelemetry 集成
- Jaeger 支持
- 控制台输出

```go
import "github.com/ishaqcherry9/depend/pkg/tracer"

// 创建追踪器
tracer := tracer.NewJaegerTracer(serviceName)
```

#### 性能分析 (Prof)
- pprof 集成
- HTTP 性能分析端点
- CPU/内存分析

#### 监控统计 (Stat)
- CPU/内存监控
- 告警通知
- 系统指标收集

### RPC 客户端

#### IM-Proxy SDK
- 完整的 IM-Proxy 服务客户端
- 用户管理、群组管理、消息管理
- 类型安全的 API 调用

```go
import "github.com/ishaqcherry9/depend/pkg/rpc/improxy"

// 初始化客户端
improxy.Init(improxy.WithBaseURL("http://localhost:8080"))

// 使用客户端
client := improxy.GetClient()
tokenResp, err := client.Register(ctx, &improxy.RegisterReq{
    UID:  "user123",
    Name: "张三",
})
```

详细文档请参考：[IM-Proxy SDK README](./pkg/rpc/improxy/README.md)

### 其他组件

- **WebSocket**：WebSocket 客户端和服务器
- **容器组**：容器管理工具
- **文件操作**：文件读写和路径处理
- **错误码**：统一的错误码定义
- **映射工具**：结构体映射和转换
- **线程池**：工作池和协程组管理
- **异常恢复**：Panic 恢复工具
- **网络工具**：IP 地址处理
- **语言工具**：多语言支持

## 📚 使用示例

### 基础使用

```go
package main

import (
    "context"
    "time"
    
    "github.com/ishaqcherry9/depend/pkg/cache"
    "github.com/ishaqcherry9/depend/pkg/logger"
    "github.com/ishaqcherry9/depend/pkg/httpcli"
)

func main() {
    ctx := context.Background()
    
    // 初始化日志
    logger.Init()
    
    // 使用缓存
    cache.Set(ctx, "key", "value", time.Hour)
    var value string
    cache.Get(ctx, "key", &value)
    
    // 发送 HTTP 请求
    resp, err := httpcli.Get(ctx, "https://api.example.com")
    if err != nil {
        logger.Error("request failed", err)
        return
    }
    
    logger.Info("response", resp)
}
```

### Gin 应用示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/ishaqcherry9/depend/pkg/gin/middleware"
    "github.com/ishaqcherry9/depend/pkg/gin/response"
)

func main() {
    router := gin.Default()
    
    // 使用中间件
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())
    router.Use(middleware.Recovery())
    
    // 路由
    router.GET("/api/user", func(c *gin.Context) {
        response.Success(c, gin.H{"user": "example"})
    })
    
    router.Run(":8080")
}
```

## 🔧 配置

大部分组件都支持通过选项模式进行配置：

```go
import "github.com/ishaqcherry9/depend/pkg/logger"

// 配置日志
logger.Init(
    logger.WithLevel("INFO"),
    logger.WithEncoding("json"),
    logger.WithOutputPath("./logs"),
)
```

## 📝 更新日志

详细的更新日志请查看 [CHANGELOG.md](./CHANGELOG.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🔗 相关链接

- [Go 官方文档](https://golang.org/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
