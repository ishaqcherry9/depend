package improxy

import (
	"context"
	"fmt"
)

// SendMsg 发送消息
func (s *client) SendMsg(ctx context.Context, req *MessageInfo) error {
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
