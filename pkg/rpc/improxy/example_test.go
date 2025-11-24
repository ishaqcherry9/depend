package improxy_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ishaqcherry9/depend/pkg/rpc/improxy"
)

func TestExampleInit(t *testing.T) {
	// 初始化全局客户端（在应用启动时调用一次）
	// baseURL 是必需的，如果未传入会 panic
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"), // 必需
		improxy.WithTimeout(30*time.Second),
	)

	// 获取全局客户端并使用
	ctx := context.Background()
	client := improxy.GetClient()
	if client == nil {
		fmt.Println("Client is not initialized")
		return
	}
	tokenResp, err := client.Register(ctx, &improxy.RegisterReq{
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

func TestExampleGetClient(t *testing.T) {
	// 必须先初始化
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)

	// 获取全局客户端
	ctx := context.Background()
	client := improxy.GetClient()
	if client == nil {
		fmt.Println("Client is not initialized")
		return
	}

	// 使用客户端
	tokenResp, err := client.Register(ctx, &improxy.RegisterReq{
		UID:  "user123",
		Name: "张三",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Token: %s\n", tokenResp.Tokens)
}

func TestExampleClient_Register(t *testing.T) {
	fmt.Print("ExampleClient_Register")
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 注册用户
	tokenResp, _ := client.Register(ctx, &improxy.RegisterReq{
		UID:  "user123",
		Name: "张三",
	})
	fmt.Printf("Token: %s\n", tokenResp.Tokens)

	// 刷新Token
	tokenResp, _ = client.RefreshToken(ctx, &improxy.RefreshReq{
		UID: "user123",
	})
	fmt.Printf("New Token: %s\n", tokenResp.Tokens)
}

func TestExampleClient_CreateTeam(t *testing.T) {
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 创建群组
	teamInfo, _ := client.CreateTeam(ctx, &improxy.CreateTeamReq{
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
	listData, _ := client.QueryTeam(ctx, &improxy.QueryTeamReq{
		TeamIds:  "123456,789012",
		TeamType: 1,
	})
	fmt.Printf("Found %d teams\n", len(listData.TeamInfoList))
}

func TestExampleClient_SendMsg(t *testing.T) {
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 发送消息
	err := client.SendMsg(ctx, &improxy.MessageInfo{
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

func TestExampleClient_GetProviderStatus(t *testing.T) {
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 获取IM渠道状态
	statusResp, _ := client.GetProviderStatus(ctx)
	fmt.Printf("Current Provider: %s\n", statusResp.CurrentProvider)
	fmt.Printf("Enabled Providers: %v\n", statusResp.EnabledProviders)
}
