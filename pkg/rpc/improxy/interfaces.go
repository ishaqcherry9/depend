package improxy

import "context"

// UserService 用户管理服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *RegisterReq) (*TokenResp, error)
	// Disable 禁用用户
	Disable(ctx context.Context, req *UIDReq) error
	// Enable 启用用户
	Enable(ctx context.Context, req *UIDReq) error
	// RefreshToken 刷新用户Token
	RefreshToken(ctx context.Context, req *RefreshReq) (*TokenResp, error)
}

// TeamService 群组管理服务接口
type TeamService interface {
	// CreateTeam 创建群组
	CreateTeam(ctx context.Context, req *CreateTeamReq) (*TeamInfoRsp, error)
	// QueryTeam 批量查询群组信息列表
	QueryTeam(ctx context.Context, req *QueryTeamReq) (*ListTeamsData, error)
	// UpdateTeam 更新群组信息
	UpdateTeam(ctx context.Context, req *UpdateTeamReq) (*UpdateTeamResp, error)
	// AddTeamMember 拉人入群
	AddTeamMember(ctx context.Context, req *AddTeamMembersReq) (*AddTeamMembersResp, error)
	// KickTeamMember 踢人出群
	KickTeamMember(ctx context.Context, req *KickTeamMembersReq) (*KickTeamMembersResp, error)
	// RemoveTeam 解散群组
	RemoveTeam(ctx context.Context, req *DeleteTeamReq) error
	// QueryTeamDetail 查询群详情
	QueryTeamDetail(ctx context.Context, req *QueryTeamDetailReq) (*QueryTeamDetailResp, error)
	// GetJoinedTeamsPaginated 查询账号已加入群组（分页）
	GetJoinedTeamsPaginated(ctx context.Context, req *GetJoinedTeamsPaginatedReq) (*GetJoinedTeamsPaginatedResp, error)
	// LeaveTeam 主动退群
	LeaveTeam(ctx context.Context, req *LeaveTeamReq) error
	// MuteTeamListAll 全员禁言/解除
	MuteTeamListAll(ctx context.Context, req *MuteAllReq) (*UpdateTeamResp, error)
	// ListTeamMembers 分页查询群成员列表
	ListTeamMembers(ctx context.Context, req *ListTeamMembersReq) (*ListTeamMembersResp, error)
}

// MessageService 消息管理服务接口
type MessageService interface {
	// SendMsg 发送消息
	SendMsg(ctx context.Context, req *MessageInfo) error
	// BroadcastMsg 发送广播消息
	BroadcastMsg(ctx context.Context, req *BroadcastMessageInfo) (*BroadcastNotificationResp, error)
}

// AuthService 认证服务接口
type AuthService interface {
	// Login 登录
	Login(ctx context.Context, req *LoginRequest) (*LoginReply, error)
	// Register 注册
	Register(ctx context.Context, req *RegisterRequest) (*RegisterReply, error)
	// Logout 登出
	Logout(ctx context.Context) error
}

// MetaService 元信息服务接口
type MetaService interface {
	// GetProviderStatus 获取当前IM渠道状态
	GetProviderStatus(ctx context.Context) (*ProviderStatusResp, error)
}

