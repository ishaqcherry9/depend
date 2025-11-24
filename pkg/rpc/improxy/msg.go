package improxy

import (
	"context"
	"fmt"
)

// SendMsg 发送消息
func (s *client) SendMsg(ctx context.Context, req *MessageInfo) error {
	// 参数校验
	if req.FromUID == "" {
		return fmt.Errorf("invalid request: from_uid is required")
	}
	if req.ToUID == "" {
		return fmt.Errorf("invalid request: to_uid is required")
	}
	if req.ToType != "1" && req.ToType != "2" {
		return fmt.Errorf("invalid request: to_type must be 1 or 2, got: %s", req.ToType)
	}
	if req.MessageType < 0 || req.MessageType > 6 {
		return fmt.Errorf("invalid request: message_type must be between 0 and 6, got: %d", req.MessageType)
	}
	if req.FromBuss == "" {
		return fmt.Errorf("invalid request: from_buss is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/msg/sendMsg", req, &resp); err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("send message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// BroadcastMsg 发送广播消息
func (s *client) BroadcastMsg(ctx context.Context, req *BroadcastMessageInfo) (*BroadcastNotificationResp, error) {
	// 参数校验
	if len(req.Content) == 0 {
		return nil, fmt.Errorf("invalid request: content cannot be empty")
	}
	if len(req.Content) > 4096 {
		return nil, fmt.Errorf("invalid request: content length cannot exceed 4096 characters, got: %d", len(req.Content))
	}
	if req.FromAccountID == "" {
		return nil, fmt.Errorf("invalid request: from_account_id is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/msg/broadcast_notification", req, &resp); err != nil {
		return nil, fmt.Errorf("broadcast message failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("broadcast message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var broadcastResp BroadcastNotificationResp
	if err := parseResponse(resp.Data, &broadcastResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &broadcastResp, nil
}
