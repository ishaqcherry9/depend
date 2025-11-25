package improxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ishaqcherry9/depend/pkg/gin/middleware"
	jsoniter "github.com/ishaqcherry9/depend/pkg/json-iterator"
	"github.com/ishaqcherry9/depend/pkg/logger"
)

var (
	gClient     Client
	gClientOnce sync.Once
)

// 约束Client接口的实现，确保client结构体实现了Client接口的所有方法
var _ Client = (*client)(nil)

// Client im-proxy SDK 客户端
type Client interface {
	// 用户管理接口
	Register(ctx context.Context, req *RegisterReq) (*TokenResp, error)
	DisableUser(ctx context.Context, req *UIDReq) error
	EnableUser(ctx context.Context, req *UIDReq) error
	RefreshToken(ctx context.Context, req *RefreshReq) (*TokenResp, error)
	QueryUserProfile(ctx context.Context, req *QueryUserProfileReq) (*QueryUserProfileResp, error)

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

	// 元信息接口
	GetProviderStatus(ctx context.Context) (*ProviderStatusResp, error)
}

type client struct {
	baseURL    string
	timeout    time.Duration
	headers    map[string]string
	httpClient *http.Client // 带连接池的 HTTP 客户端
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

// WithMaxIdleConns 设置最大空闲连接数（连接池配置）
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *client) {
		if c.httpClient != nil {
			if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
				transport.MaxIdleConns = maxIdleConns
			}
		}
	}
}

// WithMaxIdleConnsPerHost 设置每个主机最大空闲连接数（连接池配置）
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *client) {
		if c.httpClient != nil {
			if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
				transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
			}
		}
	}
}

// WithIdleConnTimeout 设置空闲连接超时时间（连接池配置）
func WithIdleConnTimeout(timeout time.Duration) Option {
	return func(c *client) {
		if c.httpClient != nil {
			if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
				transport.IdleConnTimeout = timeout
			}
		}
	}
}

// WithMaxConnsPerHost 设置每个主机最大连接数（连接池配置）
func WithMaxConnsPerHost(maxConnsPerHost int) Option {
	return func(c *client) {
		if c.httpClient != nil {
			if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
				transport.MaxConnsPerHost = maxConnsPerHost
			}
		}
	}
}

// newClient 创建客户端实例的内部函数
func newClient(opts ...Option) Client {

	c := &client{
		baseURL: "",
		timeout: 2 * time.Second,
		headers: make(map[string]string),
	}

	// 先创建带连接池的 HTTP 客户端（默认配置）
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个主机最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
		MaxConnsPerHost:     0,                // 0 表示不限制每个主机的最大连接数
	}

	c.httpClient = &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}

	// 应用所有选项
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

	// 确保 HTTP 客户端的超时时间已更新
	c.httpClient.Timeout = c.timeout

	return c
}

// request 执行 HTTP 请求的通用方法
// 使用带连接池的 HTTP 客户端，提升性能和资源利用率
func (c *client) request(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) error {
	requestURL := fmt.Sprintf("%s%s", c.baseURL, path)

	logger.Debug(ctx, "request start",
		logger.String("method", method),
		logger.String("path", path),
		logger.String("url", requestURL),
	)

	var httpReq *http.Request
	var err error

	// 构建 HTTP 请求
	if method == "GET" || method == "DELETE" {
		// GET/DELETE 请求，将 reqBody 作为查询参数
		fullURL := requestURL
		if params, ok := reqBody.(map[string]interface{}); ok && len(params) > 0 {
			// 构建查询字符串
			queryValues := url.Values{}
			for k, v := range params {
				queryValues.Add(k, fmt.Sprintf("%v", v))
			}
			queryStr := queryValues.Encode()
			if len(queryStr) > 0 {
				if strings.Contains(fullURL, "?") {
					fullURL += "&" + queryStr
				} else {
					fullURL += "?" + queryStr
				}
			}
		}
		httpReq, err = http.NewRequestWithContext(ctx, method, fullURL, nil)
	} else {
		// POST/PUT/PATCH 请求
		var bodyReader io.Reader
		if reqBody != nil {
			// 序列化请求体
			bodyBytes, marshalErr := jsoniter.Marshal(reqBody)
			if marshalErr != nil {
				logger.Error(ctx, fmt.Sprintf("marshal request body failed for %s", method),
					logger.String("method", method),
					logger.String("url", requestURL),
					logger.Err(marshalErr),
				)
				return fmt.Errorf("marshal request body failed: %w", marshalErr)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}
		httpReq, err = http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	}

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("create %s request failed", method),
			logger.String("method", method),
			logger.String("url", requestURL),
			logger.Err(err),
		)
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
		// 从ctx中获取x-request-id,并设置到请求头
		traceID := middleware.CtxRequestID(ctx)
		if traceID != "" {
			httpReq.Header.Set(middleware.ContextRequestIDKey, traceID)
		}
	}

	// 使用带连接池的 HTTP 客户端发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("%s request failed", method),
			logger.String("method", method),
			logger.String("url", requestURL),
			logger.Err(err),
		)
		return fmt.Errorf("%s request failed: %w", method, err)
	}
	defer resp.Body.Close()
	// 获取 response header 中的 traceid 并记录到日志上下文
	traceID := resp.Header.Get(middleware.ContextRequestIDKey)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		logger.Error(ctx, fmt.Sprintf("%s request failed with non-200 status", method),
			logger.String("method", method),
			logger.String("url", requestURL),
			logger.Int("status_code", resp.StatusCode),
			logger.String("response_body", string(body)),
			logger.Err(err),
			logger.String("traceid", traceID),
		)
		return errors.New(string(body))
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, "read response body failed", logger.String("method", method), logger.String("url", requestURL), logger.Err(err), logger.String("traceid", traceID))
		return fmt.Errorf("read response body failed: %w", err)
	}

	if err := jsoniter.Unmarshal(body, respBody); err != nil {
		logger.Error(ctx, "unmarshal response body failed",
			logger.String("method", method),
			logger.String("url", requestURL),
			logger.Err(err),
			logger.String("traceid", traceID),
		)
		return fmt.Errorf("unmarshal response body failed: %w", err)
	}

	logger.Info(ctx, "request success",
		logger.String("method", method),
		logger.String("url", requestURL),
		logger.Int("status_code", resp.StatusCode),
		logger.String("response_body", string(body)),
		logger.String("traceid", traceID),
	)

	return nil
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
