# im-proxy SDK

im-proxy SDK 是一个用于调用 im-proxy 服务的 Go 客户端库。它封装了 im-proxy 的所有 HTTP 接口，提供了类型安全的 RPC 调用方式。

## 功能特性

- ✅ 完整的接口覆盖：用户管理、群组管理、消息管理、认证、元信息
- ✅ 类型安全：所有请求和响应都有明确的类型定义
- ✅ 易于使用：简洁的 API 设计，支持链式调用
- ✅ 灵活配置：支持自定义超时、请求头等
- ✅ 错误处理：统一的错误处理机制

## 安装

```bash
go get github.com/ishaqcherry9/depend/pkg/rpc/improxy
```

## 快速开始

### 方式一：使用全局单例（推荐）

```go
import "github.com/ishaqcherry9/depend/pkg/rpc/improxy"

// 在应用启动时初始化全局客户端
func init() {
    improxy.Init(
        improxy.WithBaseURL("http://localhost:8080"),
        improxy.WithTimeout(30 * time.Second),
        improxy.WithHeader("Authorization", "Bearer your-token"),
    )
}

// 在代码中直接使用全局客户端
func main() {
    ctx := context.Background()
    client := improxy.GetClient()
    
    // 使用客户端
    tokenResp, err := client.User().Register(ctx, &improxy.RegisterReq{
        UID:  "user123",
        Name: "张三",
    })
}
```

### 方式二：创建独立客户端实例

```go
import "github.com/ishaqcherry9/depend/pkg/rpc/improxy"

// 创建独立的客户端实例
client := improxy.NewClient(
    improxy.WithBaseURL("http://localhost:8080"),
    improxy.WithTimeout(30 * time.Second),
)

// 或者设置自定义请求头（如认证token）
client := improxy.NewClient(
    improxy.WithBaseURL("http://localhost:8080"),
    improxy.WithHeader("Authorization", "Bearer your-token"),
)
```

### 用户管理

```go
ctx := context.Background()

// 注册用户
tokenResp, err := client.User().Register(ctx, &improxy.RegisterReq{
    UID:    "user123",
    Name:   "张三",
    Avatar: "https://example.com/avatar.jpg",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Token: %s\n", tokenResp.Tokens)

// 禁用用户
err = client.User().Disable(ctx, &improxy.UIDReq{
    UID: "user123",
})

// 启用用户
err = client.User().Enable(ctx, &improxy.UIDReq{
    UID: "user123",
})

// 刷新Token
tokenResp, err = client.User().RefreshToken(ctx, &improxy.RefreshReq{
    UID: "user123",
})
```

### 群组管理

```go
// 创建群组
teamInfo, err := client.Team().CreateTeam(ctx, &improxy.CreateTeamReq{
    OwnerAccountID: "user123",
    TeamType:       1,
    Name:           "测试群组",
    InviteAccountIDs: []string{"user456", "user789"},
    InviteMsg:      "欢迎加入群组",
    Configuration: improxy.TeamConfiguration{
        JoinMode:           0,
        AgreeMode:          1,
        InviteMode:         0,
        UpdateTeamInfoMode: 0,
    },
})

// 查询群组
listData, err := client.Team().QueryTeam(ctx, &improxy.QueryTeamReq{
    TeamIds:  "123456,789012",
    TeamType: 1,
})

// 更新群组信息
updateResp, err := client.Team().UpdateTeam(ctx, &improxy.UpdateTeamReq{
    TeamID:     "123456",
    TeamType:   1,
    Name:       "更新后的群名称",
    OperatorID: "user123",
})

// 拉人入群
addResp, err := client.Team().AddTeamMember(ctx, &improxy.AddTeamMembersReq{
    OperatorID:       "user123",
    TeamID:           "123456",
    TeamType:         1,
    InviteAccountIDs: []string{"user456", "user789"},
    Msg:              "欢迎加入群组",
})

// 踢人出群
kickResp, err := client.Team().KickTeamMember(ctx, &improxy.KickTeamMembersReq{
    OperatorID:     "user123",
    TeamID:         "123456",
    TeamType:       1,
    KickAccountIDs: []string{"user456"},
})

// 查询群详情
detailResp, err := client.Team().QueryTeamDetail(ctx, &improxy.QueryTeamDetailReq{
    TeamID:   "123456",
    TeamType: 1,
})

// 查询已加入的群组（分页）
teamsResp, err := client.Team().GetJoinedTeamsPaginated(ctx, &improxy.GetJoinedTeamsPaginatedReq{
    AccountID: "user123",
    TeamType:  1,
    Limit:     20,
})

// 主动退群
err = client.Team().LeaveTeam(ctx, &improxy.LeaveTeamReq{
    AccountID: "user123",
    TeamID:    "123456",
    TeamType:  1,
})

// 全员禁言
muteResp, err := client.Team().MuteTeamListAll(ctx, &improxy.MuteAllReq{
    TeamID:         "123456",
    TeamType:       1,
    OperatorID:     "user123",
    ChatBannedMode: 1,
})

// 分页查询群成员列表
membersResp, err := client.Team().ListTeamMembers(ctx, &improxy.ListTeamMembersReq{
    TeamID:     "123456",
    TeamType:   1,
    Descending: false,
    Limit:      20,
})

// 解散群组
err = client.Team().RemoveTeam(ctx, &improxy.DeleteTeamReq{
    TeamID:     "123456",
    TeamType:   1,
    OperatorID: "user123",
})
```

### 消息管理

```go
// 发送消息
err := client.Message().SendMsg(ctx, &improxy.MessageInfo{
    FromUID:     "user123",
    ToUID:       "user456",
    ToType:      "1", // 1：单聊会话；2：高级群会话
    MessageType: 0,   // 0：文本消息
    Text:        "Hello World",
    FromBuss:    "business",
})

// 发送广播消息
broadcastResp, err := client.Message().BroadcastMsg(ctx, &improxy.BroadcastMessageInfo{
    Content:       "系统通知：服务器将于今晚进行维护",
    FromAccountID: "system",
    TTL:           168, // 7天
    TargetOS:      []string{"ios", "aos"},
})
```

### 认证

```go
// 登录
loginResp, err := client.Auth().Login(ctx, &improxy.LoginRequest{
    Username: "admin",
    Password: "123456",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Token: %s\n", loginResp.Data.Token)

// 注册
registerResp, err := client.Auth().Register(ctx, &improxy.RegisterRequest{
    Username: "newuser",
    Password: "password123",
    Email:    "user@example.com",
})

// 登出
err = client.Auth().Logout(ctx)
```

### 元信息

```go
// 获取IM渠道状态
statusResp, err := client.Meta().GetProviderStatus(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Current Provider: %s\n", statusResp.CurrentProvider)
fmt.Printf("Enabled Providers: %v\n", statusResp.EnabledProviders)
```

## 配置选项

### WithBaseURL

设置 im-proxy 服务的基础 URL。

```go
// 全局单例方式
improxy.Init(
    improxy.WithBaseURL("https://api.example.com"),
)

// 独立实例方式
client := improxy.NewClient(
    improxy.WithBaseURL("https://api.example.com"),
)
```

### WithTimeout

设置请求超时时间。

```go
// 全局单例方式
improxy.Init(
    improxy.WithTimeout(60 * time.Second),
)

// 独立实例方式
client := improxy.NewClient(
    improxy.WithTimeout(60 * time.Second),
)
```

### WithHeader

设置单个请求头。

```go
// 全局单例方式
improxy.Init(
    improxy.WithHeader("Authorization", "Bearer token"),
    improxy.WithHeader("X-Request-ID", "request-id"),
)

// 独立实例方式
client := improxy.NewClient(
    improxy.WithHeader("Authorization", "Bearer token"),
    improxy.WithHeader("X-Request-ID", "request-id"),
)
```

### WithHeaders

批量设置请求头。

```go
// 全局单例方式
improxy.Init(
    improxy.WithHeaders(map[string]string{
        "Authorization": "Bearer token",
        "X-Request-ID":  "request-id",
    }),
)

// 独立实例方式
client := improxy.NewClient(
    improxy.WithHeaders(map[string]string{
        "Authorization": "Bearer token",
        "X-Request-ID":  "request-id",
    }),
)
```

## 全局单例 API

### Init

初始化全局 im-proxy SDK 客户端。应该在应用启动时调用一次。

```go
improxy.Init(
    improxy.WithBaseURL("http://localhost:8080"),
    improxy.WithTimeout(30 * time.Second),
)
```

**注意：**
- `Init` 使用 `sync.Once` 确保只初始化一次
- 如果未调用 `Init`，`GetClient()` 会返回 `nil`
- **`baseURL` 是必需的**，如果未传入会 panic
- 如果 `timeout` 为 0，默认使用 `30 * time.Second`

### GetClient

获取全局 im-proxy SDK 客户端。

```go
client := improxy.GetClient()
if client == nil {
    log.Fatal("improxy client is not initialized, please call improxy.Init() first")
}
tokenResp, err := client.User().Register(ctx, &improxy.RegisterReq{
    UID:  "user123",
    Name: "张三",
})
```

**注意：**
- 如果未通过 `Init()` 初始化，`GetClient()` 会返回 `nil`
- 使用前必须先调用 `Init()` 进行初始化
- 线程安全，可以在多个 goroutine 中并发调用

## 错误处理

所有方法都会返回错误，建议使用标准的 Go 错误处理方式：

```go
tokenResp, err := client.User().Register(ctx, &improxy.RegisterReq{
    UID:  "user123",
    Name: "张三",
})
if err != nil {
    // 处理错误
    log.Printf("Register failed: %v", err)
    return err
}
// 使用 tokenResp
```

## 接口列表

### 用户管理 (UserService)

- `Register` - 用户注册
- `Disable` - 禁用用户
- `Enable` - 启用用户
- `RefreshToken` - 刷新用户Token

### 群组管理 (TeamService)

- `CreateTeam` - 创建群组
- `QueryTeam` - 批量查询群组信息列表
- `UpdateTeam` - 更新群组信息
- `AddTeamMember` - 拉人入群
- `KickTeamMember` - 踢人出群
- `RemoveTeam` - 解散群组
- `QueryTeamDetail` - 查询群详情
- `GetJoinedTeamsPaginated` - 查询账号已加入群组（分页）
- `LeaveTeam` - 主动退群
- `MuteTeamListAll` - 全员禁言/解除
- `ListTeamMembers` - 分页查询群成员列表

### 消息管理 (MessageService)

- `SendMsg` - 发送消息
- `BroadcastMsg` - 发送广播消息

### 认证 (AuthService)

- `Login` - 登录
- `Register` - 注册
- `Logout` - 登出

### 元信息 (MetaService)

- `GetProviderStatus` - 获取当前IM渠道状态

## 注意事项

1. 所有接口都需要传入 `context.Context`，用于控制请求的取消和超时
2. 请求和响应类型都有详细的字段注释，请参考类型定义
3. SDK 会自动处理 JSON 序列化和反序列化
4. 所有错误都会包含详细的错误信息，便于调试

## License

MIT

