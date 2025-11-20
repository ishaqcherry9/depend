package improxy_test

import (
	"context"
	"fmt"
	"time"

	"github.com/ishaqcherry9/depend/pkg/rpc/improxy"
)

func ExampleInit() {
	// 初始化全局客户端（在应用启动时调用一次）
	// baseURL 是必需的，如果未传入会 panic
	improxy.Init(
		improxy.WithBaseURL("http://localhost:8080"), // 必需
		improxy.WithTimeout(30*time.Second),
	)

	// 获取全局客户端并使用
	ctx := context.Background()
	client := improxy.GetClient()
	if client == nil {
		fmt.Println("Client is not initialized")
		return
	}
	tokenResp, err := client.User().Register(ctx, &improxy.RegisterReq{
		UID:    "user123",
		Name:   "张三",
		Avatar: "https://example.com/avatar.jpg",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Token: %s\n", tokenResp.Tokens)
}

func ExampleGetClient() {
	// 必须先初始化
	improxy.Init(
		improxy.WithBaseURL("http://localhost:8080"),
	)

	// 获取全局客户端
	ctx := context.Background()
	client := improxy.GetClient()
	if client == nil {
		fmt.Println("Client is not initialized")
		return
	}

	// 使用客户端
	tokenResp, err := client.User().Register(ctx, &improxy.RegisterReq{
		UID:  "user123",
		Name: "张三",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Token: %s\n", tokenResp.Tokens)
}

func ExampleNewClient() {
	// 创建独立的客户端实例
	client := improxy.NewClient(
		improxy.WithBaseURL("http://localhost:8080"),
		improxy.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// 使用用户管理接口
	userService := client.User()
	tokenResp, err := userService.Register(ctx, &improxy.RegisterReq{
		UID:    "user123",
		Name:   "张三",
		Avatar: "https://example.com/avatar.jpg",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Token: %s\n", tokenResp.Tokens)
}

func ExampleUserService() {
	client := improxy.NewClient(
		improxy.WithBaseURL("http://localhost:8080"),
	)

	ctx := context.Background()

	// 注册用户
	tokenResp, _ := client.User().Register(ctx, &improxy.RegisterReq{
		UID:  "user123",
		Name: "张三",
	})
	fmt.Printf("Token: %s\n", tokenResp.Tokens)

	// 刷新Token
	tokenResp, _ = client.User().RefreshToken(ctx, &improxy.RefreshReq{
		UID: "user123",
	})
	fmt.Printf("New Token: %s\n", tokenResp.Tokens)
}

func ExampleTeamService() {
	client := improxy.NewClient(
		improxy.WithBaseURL("http://localhost:8080"),
	)

	ctx := context.Background()

	// 创建群组
	teamInfo, _ := client.Team().CreateTeam(ctx, &improxy.CreateTeamReq{
		OwnerAccountID:   "user123",
		TeamType:         1,
		Name:             "测试群组",
		InviteAccountIDs: []string{"user456", "user789"},
		InviteMsg:        "欢迎加入群组",
		Configuration: improxy.TeamConfiguration{
			JoinMode:           0,
			AgreeMode:          1,
			InviteMode:         0,
			UpdateTeamInfoMode: 0,
		},
	})
	fmt.Printf("Team ID: %s\n", teamInfo.TeamID)

	// 查询群组
	listData, _ := client.Team().QueryTeam(ctx, &improxy.QueryTeamReq{
		TeamIds:  "123456,789012",
		TeamType: 1,
	})
	fmt.Printf("Found %d teams\n", len(listData.TeamInfoList))
}

func ExampleMessageService() {
	client := improxy.NewClient(
		improxy.WithBaseURL("http://localhost:8080"),
	)

	ctx := context.Background()

	// 发送消息
	err := client.Message().SendMsg(ctx, &improxy.MessageInfo{
		FromUID:     "user123",
		ToUID:       "user456",
		ToType:      "1",
		MessageType: 0,
		Text:        "Hello World",
		FromBuss:    "business",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Message sent successfully")
}

func ExampleAuthService() {
	client := improxy.NewClient(
		improxy.WithBaseURL("http://localhost:8080"),
	)

	ctx := context.Background()

	// 登录
	loginResp, _ := client.Auth().Login(ctx, &improxy.LoginRequest{
		Username: "admin",
		Password: "123456",
	})
	fmt.Printf("Token: %s\n", loginResp.Data.Token)
}

func ExampleMetaService() {
	client := improxy.NewClient(
		improxy.WithBaseURL("http://localhost:8080"),
	)

	ctx := context.Background()

	// 获取IM渠道状态
	statusResp, _ := client.Meta().GetProviderStatus(ctx)
	fmt.Printf("Current Provider: %s\n", statusResp.CurrentProvider)
	fmt.Printf("Enabled Providers: %v\n", statusResp.EnabledProviders)
}

