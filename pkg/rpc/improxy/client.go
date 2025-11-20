package improxy

import (
	"context"
	"fmt"
	"time"

	"github.com/ishaqcherry9/depend/pkg/httpcli"
	jsoniter "github.com/ishaqcherry9/depend/pkg/json-iterator"
)

// Client im-proxy SDK 客户端
type Client interface {
	// User 用户管理接口
	User() UserService
	// Team 群组管理接口
	Team() TeamService
	// Message 消息管理接口
	Message() MessageService
	// Auth 认证接口
	Auth() AuthService
	// Meta 元信息接口
	Meta() MetaService
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

// NewClient 创建新的 im-proxy SDK 客户端
func NewClient(opts ...Option) Client {
	c := &client{
		baseURL: "http://localhost:8080",
		timeout: 30 * time.Second,
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	// 设置默认 Content-Type
	if _, ok := c.headers["Content-Type"]; !ok {
		c.headers["Content-Type"] = "application/json"
	}

	return c
}

// User 返回用户管理服务
func (c *client) User() UserService {
	return &userService{client: c}
}

// Team 返回群组管理服务
func (c *client) Team() TeamService {
	return &teamService{client: c}
}

// Message 返回消息管理服务
func (c *client) Message() MessageService {
	return &messageService{client: c}
}

// Auth 返回认证服务
func (c *client) Auth() AuthService {
	return &authService{client: c}
}

// Meta 返回元信息服务
func (c *client) Meta() MetaService {
	return &metaService{client: c}
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
