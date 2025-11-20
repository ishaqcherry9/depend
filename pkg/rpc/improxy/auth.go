package improxy

import (
	"context"
	"fmt"
)

type authService struct {
	client *client
}

// Login 登录
func (s *authService) Login(ctx context.Context, req *LoginRequest) (*LoginReply, error) {
	var resp StandardResponse
	if err := s.client.request(ctx, "POST", "/api/v1/auth/login", req, &resp); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("login failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var loginResp LoginReply
	if err := parseResponse(resp.Data, &loginResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	loginResp.Code = resp.Code
	loginResp.Msg = resp.Msg

	return &loginResp, nil
}

// Register 注册
func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*RegisterReply, error) {
	var resp StandardResponse
	if err := s.client.request(ctx, "POST", "/api/v1/auth/register", req, &resp); err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("register failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var registerResp RegisterReply
	if err := parseResponse(resp.Data, &registerResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	registerResp.Code = resp.Code
	registerResp.Msg = resp.Msg

	return &registerResp, nil
}

// Logout 登出
func (s *authService) Logout(ctx context.Context) error {
	var resp StandardResponse
	if err := s.client.request(ctx, "POST", "/api/v1/auth/logout", nil, &resp); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("logout failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

