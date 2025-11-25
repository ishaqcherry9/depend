package improxy

import (
	"context"
	"fmt"
)

// Register 用户注册
func (s *client) Register(ctx context.Context, req *RegisterReq) (*TokenResp, error) {
	// 参数校验
	if req.UID == "" {
		return nil, fmt.Errorf("invalid request: uid is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/user/register", req, &resp); err != nil {
		return nil, fmt.Errorf("register user failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("register user failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var tokenResp TokenResp
	if err := parseResponse(resp.Data, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &tokenResp, nil
}

// Disable 禁用用户
func (s *client) DisableUser(ctx context.Context, req *UIDReq) error {
	// 参数校验
	if req.UID == "" {
		return fmt.Errorf("invalid request: uid is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/user/disable", req, &resp); err != nil {
		return fmt.Errorf("disable user failed: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("disable user failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// Enable 启用用户
func (s *client) EnableUser(ctx context.Context, req *UIDReq) error {
	// 参数校验
	if req.UID == "" {
		return fmt.Errorf("invalid request: uid is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/user/enable", req, &resp); err != nil {
		return fmt.Errorf("enable user failed: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("enable user failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// RefreshToken 刷新用户Token
func (s *client) RefreshToken(ctx context.Context, req *RefreshReq) (*TokenResp, error) {
	// 参数校验
	if req.UID == "" {
		return nil, fmt.Errorf("invalid request: uid is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/user/refreshToken", req, &resp); err != nil {
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("refresh token failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var tokenResp TokenResp
	if err := parseResponse(resp.Data, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &tokenResp, nil
}

// QueryUserProfile 查询用户名片
func (s *client) QueryUserProfile(ctx context.Context, req *QueryUserProfileReq) (*QueryUserProfileResp, error) {
	// 参数校验
	if req.UID == "" {
		return nil, fmt.Errorf("invalid request: uid is required")
	}

	params := map[string]interface{}{
		"uid": req.UID,
	}

	var resp StandardResponse
	if err := s.request(ctx, "GET", "/api/v1/im/user/profile", params, &resp); err != nil {
		return nil, fmt.Errorf("query user profile failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("query user profile failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var profileResp QueryUserProfileResp
	if err := parseResponse(resp.Data, &profileResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &profileResp, nil
}
