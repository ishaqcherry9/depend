package improxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ishaqcherry9/depend/pkg/httpcli"
	jsoniter "github.com/ishaqcherry9/depend/pkg/json-iterator"
)

var (
	gClient     Client
	gClientOnce sync.Once
)

// Client im-proxy SDK 客户端
type Client interface {
	// 用户管理接口
	Register(ctx context.Context, req *RegisterReq) (*TokenResp, error)
	DisableUser(ctx context.Context, req *UIDReq) error
	EnableUser(ctx context.Context, req *UIDReq) error
	RefreshToken(ctx context.Context, req *RefreshReq) (*TokenResp, error)

	// 群组管理接口
	CreateTeam(ctx context.Context, req *CreateTeamReq) (*TeamInfoRsp, error)
	QueryTeam(ctx context.Context, req *QueryTeamReq) (*ListTeamsData, error)
	UpdateTeam(ctx context.Context, req *UpdateTeamReq) (*UpdateTeamResp, error)
	AddTeamMember(ctx context.Context, req *AddTeamMembersReq) (*AddTeamMembersResp, error)
	KickTeamMember(ctx context.Context, req *KickTeamMembersReq) (*KickTeamMembersResp, error)
	RemoveTeam(ctx context.Context, req *DeleteTeamReq) error
	QueryTeamDetail(ctx context.Context, req *QueryTeamDetailReq) (*QueryTeamDetailResp, error)
	GetJoinedTeamsPaginated(ctx context.Context, req *GetJoinedTeamsPaginatedReq) (*GetJoinedTeamsPaginatedResp, error)
	LeaveTeam(ctx context.Context, req *LeaveTeamReq) error
	MuteTeamListAll(ctx context.Context, req *MuteAllReq) (*UpdateTeamResp, error)
	ListTeamMembers(ctx context.Context, req *ListTeamMembersReq) (*ListTeamMembersResp, error)

	// 消息管理接口
	SendMsg(ctx context.Context, req *MessageInfo) error
	BroadcastMsg(ctx context.Context, req *BroadcastMessageInfo) (*BroadcastNotificationResp, error)

	// 认证接口
	Login(ctx context.Context, req *LoginRequest) (*LoginReply, error)
	RegisterAuth(ctx context.Context, req *RegisterRequest) (*RegisterReply, error)
	Logout(ctx context.Context) error

	// 元信息接口
	GetProviderStatus(ctx context.Context) (*ProviderStatusResp, error)
}

type client struct {
	baseURL string
	timeout time.Duration
	headers map[string]string
}

// Option 客户端配置选项
type Option func(*client)

// WithBaseURL 设置基础URL
func WithBaseURL(url string) Option {
	return func(c *client) {
		c.baseURL = url
	}
}

// WithTimeout 设置请求超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *client) {
		c.timeout = timeout
	}
}

// WithHeaders 设置默认请求头
func WithHeaders(headers map[string]string) Option {
	return func(c *client) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithHeader 设置单个请求头
func WithHeader(key, value string) Option {
	return func(c *client) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		c.headers[key] = value
	}
}

// newClient 创建客户端实例的内部函数
func newClient(opts ...Option) Client {

	c := &client{
		baseURL: "",
		timeout: 2 * time.Second,
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	// 检查 baseURL
	if c.baseURL == "" {
		panic("improxy: baseURL is required, please use improxy.WithBaseURL() to set it")
	}

	// 设置默认 Content-Type
	if _, ok := c.headers["Content-Type"]; !ok {
		c.headers["Content-Type"] = "application/json"
	}

	return c
}

// 用户管理接口实现
func (c *client) Register(ctx context.Context, req *RegisterReq) (*TokenResp, error) {
	return (&userService{client: c}).Register(ctx, req)
}

func (c *client) DisableUser(ctx context.Context, req *UIDReq) error {
	return (&userService{client: c}).Disable(ctx, req)
}

func (c *client) EnableUser(ctx context.Context, req *UIDReq) error {
	return (&userService{client: c}).Enable(ctx, req)
}

func (c *client) RefreshToken(ctx context.Context, req *RefreshReq) (*TokenResp, error) {
	return (&userService{client: c}).RefreshToken(ctx, req)
}

// 群组管理接口实现
func (c *client) CreateTeam(ctx context.Context, req *CreateTeamReq) (*TeamInfoRsp, error) {
	return (&teamService{client: c}).CreateTeam(ctx, req)
}

func (c *client) QueryTeam(ctx context.Context, req *QueryTeamReq) (*ListTeamsData, error) {
	return (&teamService{client: c}).QueryTeam(ctx, req)
}

func (c *client) UpdateTeam(ctx context.Context, req *UpdateTeamReq) (*UpdateTeamResp, error) {
	return (&teamService{client: c}).UpdateTeam(ctx, req)
}

func (c *client) AddTeamMember(ctx context.Context, req *AddTeamMembersReq) (*AddTeamMembersResp, error) {
	return (&teamService{client: c}).AddTeamMember(ctx, req)
}

func (c *client) KickTeamMember(ctx context.Context, req *KickTeamMembersReq) (*KickTeamMembersResp, error) {
	return (&teamService{client: c}).KickTeamMember(ctx, req)
}

func (c *client) RemoveTeam(ctx context.Context, req *DeleteTeamReq) error {
	return (&teamService{client: c}).RemoveTeam(ctx, req)
}

func (c *client) QueryTeamDetail(ctx context.Context, req *QueryTeamDetailReq) (*QueryTeamDetailResp, error) {
	return (&teamService{client: c}).QueryTeamDetail(ctx, req)
}

func (c *client) GetJoinedTeamsPaginated(ctx context.Context, req *GetJoinedTeamsPaginatedReq) (*GetJoinedTeamsPaginatedResp, error) {
	return (&teamService{client: c}).GetJoinedTeamsPaginated(ctx, req)
}

func (c *client) LeaveTeam(ctx context.Context, req *LeaveTeamReq) error {
	return (&teamService{client: c}).LeaveTeam(ctx, req)
}

func (c *client) MuteTeamListAll(ctx context.Context, req *MuteAllReq) (*UpdateTeamResp, error) {
	return (&teamService{client: c}).MuteTeamListAll(ctx, req)
}

func (c *client) ListTeamMembers(ctx context.Context, req *ListTeamMembersReq) (*ListTeamMembersResp, error) {
	return (&teamService{client: c}).ListTeamMembers(ctx, req)
}

// 消息管理接口实现
func (c *client) SendMsg(ctx context.Context, req *MessageInfo) error {
	return (&messageService{client: c}).SendMsg(ctx, req)
}

func (c *client) BroadcastMsg(ctx context.Context, req *BroadcastMessageInfo) (*BroadcastNotificationResp, error) {
	return (&messageService{client: c}).BroadcastMsg(ctx, req)
}

// 认证接口实现
func (c *client) Login(ctx context.Context, req *LoginRequest) (*LoginReply, error) {
	return (&authService{client: c}).Login(ctx, req)
}

func (c *client) RegisterAuth(ctx context.Context, req *RegisterRequest) (*RegisterReply, error) {
	return (&authService{client: c}).Register(ctx, req)
}

func (c *client) Logout(ctx context.Context) error {
	return (&authService{client: c}).Logout(ctx)
}

// 元信息接口实现
func (c *client) GetProviderStatus(ctx context.Context) (*ProviderStatusResp, error) {
	return (&metaService{client: c}).GetProviderStatus(ctx)
}

// request 执行 HTTP 请求的通用方法
func (c *client) request(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) error {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req := httpcli.New()
	req.SetURL(url)
	req.SetTimeout(c.timeout)

	// 设置请求头
	for k, v := range c.headers {
		req.SetHeader(k, v)
	}

	// 如果是 GET 请求，将 reqBody 作为查询参数
	if method == "GET" {
		if params, ok := reqBody.(map[string]interface{}); ok {
			req.SetParams(params)
		}
		resp, err := req.GET()
		if err != nil {
			return fmt.Errorf("GET request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := resp.ReadBody()
			return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}

		return resp.BindJSON(respBody)
	}

	// POST/PUT/PATCH/DELETE 请求
	if reqBody != nil {
		req.SetBody(reqBody)
	}
	var resp *httpcli.Response
	var err error

	switch method {
	case "POST":
		resp, err = req.POST()
	case "PUT":
		resp, err = req.PUT()
	case "PATCH":
		resp, err = req.PATCH()
	case "DELETE":
		resp, err = req.DELETE()
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("%s request failed: %w", method, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := resp.ReadBody()
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return resp.BindJSON(respBody)
}

// StandardResponse 标准响应格式
type StandardResponse struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId,omitempty"`
}

// parseResponse 解析标准响应
func parseResponse(data interface{}, target interface{}) error {
	// 将 data 转换为 JSON 再解析到 target
	jsonBytes, err := jsoniter.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal response data failed: %w", err)
	}

	if err := jsoniter.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("unmarshal response data failed: %w", err)
	}

	return nil
}

// Init 初始化全局 im-proxy SDK 客户端
// baseURL 必须传入，否则会 panic
func Init(opts ...Option) Client {
	gClientOnce.Do(func() {
		gClient = newClient(opts...)
	})
	return gClient
}

// GetClient 获取全局 im-proxy SDK 客户端
func GetClient() Client {
	return gClient
}
