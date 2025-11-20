package improxy

import (
	"context"
	"fmt"
)

type metaService struct {
	client *client
}

// GetProviderStatus 获取当前IM渠道状态
func (s *metaService) GetProviderStatus(ctx context.Context) (*ProviderStatusResp, error) {
	var resp StandardResponse
	if err := s.client.request(ctx, "GET", "/api/v1/im/meta/providerStatus", nil, &resp); err != nil {
		return nil, fmt.Errorf("get provider status failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("get provider status failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var statusResp ProviderStatusResp
	if err := parseResponse(resp.Data, &statusResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	statusResp.Code = resp.Code
	statusResp.Msg = resp.Msg

	return &statusResp, nil
}

