package improxy

import (
	"context"
	"fmt"
)

// CreateTeam 创建群组
func (s *client) CreateTeam(ctx context.Context, req *CreateTeamReq) (*TeamInfoRsp, error) {
	// 参数校验
	if req.OwnerAccountID == "" {
		return nil, fmt.Errorf("invalid request: owner_account_id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("invalid request: name is required")
	}
	if req.TeamType != 0 {
		if req.TeamType != 1 && req.TeamType != 2 {
			return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
		}
	} else {
		// 默认值为 1
		req.TeamType = 1
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/create", req, &resp); err != nil {
		return nil, fmt.Errorf("create team failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("create team failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var teamInfo TeamInfoRsp
	if err := parseResponse(resp.Data, &teamInfo); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &teamInfo, nil
}

// QueryTeam 批量查询群组信息列表
func (s *client) QueryTeam(ctx context.Context, req *QueryTeamReq) (*ListTeamsData, error) {
	// 参数校验
	if req.TeamIds == "" {
		return nil, fmt.Errorf("invalid request: team_ids is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}

	params := map[string]interface{}{
		"team_ids":  req.TeamIds,
		"team_type": req.TeamType,
	}

	var resp StandardResponse
	if err := s.request(ctx, "GET", "/api/v1/im/team/query", params, &resp); err != nil {
		return nil, fmt.Errorf("query team failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("query team failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var listData ListTeamsData
	if err := parseResponse(resp.Data, &listData); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &listData, nil
}

// UpdateTeam 更新群组信息
func (s *client) UpdateTeam(ctx context.Context, req *UpdateTeamReq) (*UpdateTeamResp, error) {
	// 参数校验
	if req.TeamID == "" {
		return nil, fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}
	if req.OperatorID == "" {
		return nil, fmt.Errorf("invalid request: operator_id is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/update", req, &resp); err != nil {
		return nil, fmt.Errorf("update team failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("update team failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var updateResp UpdateTeamResp
	if err := parseResponse(resp.Data, &updateResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &updateResp, nil
}

// AddTeamMember 拉人入群
func (s *client) AddTeamMember(ctx context.Context, req *AddTeamMembersReq) (*AddTeamMembersResp, error) {
	// 参数校验
	if req.OperatorID == "" {
		return nil, fmt.Errorf("invalid request: operator_id is required")
	}
	if req.TeamID == "" {
		return nil, fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}
	if len(req.InviteAccountIDs) == 0 {
		return nil, fmt.Errorf("invalid request: invite_account_ids cannot be empty")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/add", req, &resp); err != nil {
		return nil, fmt.Errorf("add team member failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("add team member failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var addResp AddTeamMembersResp
	if err := parseResponse(resp.Data, &addResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &addResp, nil
}

// KickTeamMember 踢人出群
func (s *client) KickTeamMember(ctx context.Context, req *KickTeamMembersReq) (*KickTeamMembersResp, error) {
	// 参数校验
	if req.OperatorID == "" {
		return nil, fmt.Errorf("invalid request: operator_id is required")
	}
	if req.TeamID == "" {
		return nil, fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}
	// 踢出账号ID列表，最多10个，最少1个
	if len(req.KickAccountIDs) < 1 || len(req.KickAccountIDs) > 10 {
		return nil, fmt.Errorf("invalid request: kick_account_ids length must be between 1 and 10, got: %d", len(req.KickAccountIDs))
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/kick", req, &resp); err != nil {
		return nil, fmt.Errorf("kick team member failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("kick team member failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var kickResp KickTeamMembersResp
	if err := parseResponse(resp.Data, &kickResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &kickResp, nil
}

// RemoveTeam 解散群组
func (s *client) RemoveTeam(ctx context.Context, req *DeleteTeamReq) error {
	// 参数校验
	if req.TeamID == "" {
		return fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}
	if req.OperatorID == "" {
		return fmt.Errorf("invalid request: operator_id is required")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/remove", req, &resp); err != nil {
		return fmt.Errorf("remove team failed: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("remove team failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// QueryTeamDetail 查询群详情
func (s *client) QueryTeamDetail(ctx context.Context, req *QueryTeamDetailReq) (*QueryTeamDetailResp, error) {
	// 参数校验
	if req.TeamID == "" {
		return nil, fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}

	params := map[string]interface{}{
		"team_id":   req.TeamID,
		"team_type": req.TeamType,
	}

	var resp StandardResponse
	if err := s.request(ctx, "GET", "/api/v1/im/team/queryDetail", params, &resp); err != nil {
		return nil, fmt.Errorf("query team detail failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("query team detail failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var detailResp QueryTeamDetailResp
	if err := parseResponse(resp.Data, &detailResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &detailResp, nil
}

// GetJoinedTeamsPaginated 查询账号已加入群组（分页）
func (s *client) GetJoinedTeamsPaginated(ctx context.Context, req *GetJoinedTeamsPaginatedReq) (*GetJoinedTeamsPaginatedResp, error) {
	// 参数校验
	if req.AccountID == "" {
		return nil, fmt.Errorf("invalid request: account_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}

	params := map[string]interface{}{
		"account_id": req.AccountID,
		"team_type":  req.TeamType,
	}
	if req.PageToken != "" {
		params["page_token"] = req.PageToken
	}
	// 如果 limit <= 0，设置默认值为 10（与 im-proxy 保持一致）
	if req.Limit <= 0 {
		req.Limit = 10
	}
	params["limit"] = req.Limit

	var resp StandardResponse
	if err := s.request(ctx, "GET", "/api/v1/im/team/joinTeams", params, &resp); err != nil {
		return nil, fmt.Errorf("get joined teams failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("get joined teams failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var teamsResp GetJoinedTeamsPaginatedResp
	if err := parseResponse(resp.Data, &teamsResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &teamsResp, nil
}

// LeaveTeam 主动退群
func (s *client) LeaveTeam(ctx context.Context, req *LeaveTeamReq) error {
	// 参数校验
	if req.AccountID == "" {
		return fmt.Errorf("invalid request: account_id is required")
	}
	if req.TeamID == "" {
		return fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/leave", req, &resp); err != nil {
		return fmt.Errorf("leave team failed: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("leave team failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// MuteTeamListAll 全员禁言/解除
func (s *client) MuteTeamListAll(ctx context.Context, req *MuteAllReq) (*UpdateTeamResp, error) {
	// 参数校验
	if req.TeamID == "" {
		return nil, fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}
	if req.OperatorID == "" {
		return nil, fmt.Errorf("invalid request: operator_id is required")
	}
	// chat_banned_mode: 0-取消禁言，1-禁言普通成员，3-禁言全体成员
	if req.ChatBannedMode < 0 || req.ChatBannedMode > 3 {
		return nil, fmt.Errorf("invalid request: chat_banned_mode must be between 0 and 3, got: %d", req.ChatBannedMode)
	}
	// chat_banned_mode 不能是 2
	if req.ChatBannedMode == 2 {
		return nil, fmt.Errorf("invalid request: chat_banned_mode cannot be 2")
	}

	var resp StandardResponse
	if err := s.request(ctx, "POST", "/api/v1/im/team/muteTlistAll", req, &resp); err != nil {
		return nil, fmt.Errorf("mute team list all failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("mute team list all failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var muteResp UpdateTeamResp
	if err := parseResponse(resp.Data, &muteResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &muteResp, nil
}

// ListTeamMembers 分页查询群成员列表
func (s *client) ListTeamMembers(ctx context.Context, req *ListTeamMembersReq) (*ListTeamMembersResp, error) {
	// 参数校验
	if req.TeamID == "" {
		return nil, fmt.Errorf("invalid request: team_id is required")
	}
	if req.TeamType == 0 {
		return nil, fmt.Errorf("invalid request: team_type is required")
	}
	if req.TeamType != 1 && req.TeamType != 2 {
		return nil, fmt.Errorf("invalid request: team_type must be 1 or 2, got: %d", req.TeamType)
	}
	// limit 最大为 100
	if req.Limit > 0 && req.Limit > 100 {
		return nil, fmt.Errorf("invalid request: limit cannot exceed 100, got: %d", req.Limit)
	}

	params := map[string]interface{}{
		"team_id":   req.TeamID,
		"team_type": req.TeamType,
	}
	if req.Descending {
		params["descending"] = true
	}
	if req.PageToken != "" {
		params["page_token"] = req.PageToken
	}
	if req.Limit > 0 {
		params["limit"] = req.Limit
	}

	var resp StandardResponse
	if err := s.request(ctx, "GET", "/api/v1/im/team/listMembers", params, &resp); err != nil {
		return nil, fmt.Errorf("list team members failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("list team members failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var membersResp ListTeamMembersResp
	if err := parseResponse(resp.Data, &membersResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &membersResp, nil
}
