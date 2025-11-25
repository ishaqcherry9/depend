package improxy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ishaqcherry9/depend/pkg/rpc/improxy"
)

// 测试注册
func TestExampleClient_Register(t *testing.T) {
	fmt.Print("ExampleClient_Register")
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 注册用户
	tokenResp, err := client.Register(ctx, &improxy.RegisterReq{
		UID:  "user1",
		Name: "张三",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Token: %s\n", tokenResp.Tokens)
}

func TestExampleClient_RefreshToken(t *testing.T) {
	fmt.Print("ExampleClient_RefreshToken")
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 刷新Token
	tokenResp, err := client.RefreshToken(ctx, &improxy.RefreshReq{
		UID: "user123",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("New Token: %s\n", tokenResp.Tokens)
}

// 测试查询用户名片
func TestExampleClient_QueryUserProfile(t *testing.T) {
	fmt.Print("ExampleClient_QueryUserProfile")
	// 初始化客户端
	improxy.Init(
		improxy.WithBaseURL("http://localhost:9080"),
	)
	client := improxy.GetClient()

	ctx := context.Background()

	// 查询用户名片
	userProfile, err := client.QueryUserProfile(ctx, &improxy.QueryUserProfileReq{
		UID: "user1",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("User Profile: %v\n", userProfile)
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
