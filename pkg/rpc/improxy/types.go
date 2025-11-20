package improxy

// ========== 用户管理相关类型 ==========

// RegisterReq 用户注册请求
type RegisterReq struct {
	UID    string `json:"uid"`    // 用户唯一ID
	Name   string `json:"name"`   // 用户昵称
	Avatar string `json:"avatar"` // 用户头像
}

// TokenResp 返回tokens
type TokenResp struct {
	Tokens string `json:"tokens"` // 令牌
}

// UIDReq 仅含UID
type UIDReq struct {
	UID string `json:"uid"` // 用户唯一ID
}

// RefreshReq 刷新token请求
type RefreshReq struct {
	UID string `json:"uid"` // 用户唯一ID
}

// ========== 群组管理相关类型 ==========

// CreateTeamReq 创建群组请求
type CreateTeamReq struct {
	OwnerAccountID   string            `json:"owner_account_id"`   // 群主账号ID
	TeamType         int               `json:"team_type"`          // 群组类型,默认为1,云信支持高级群(team_type=1)和超大群(team_type=2)两种类型
	Name             string            `json:"name"`               // 群名称
	Icon             string            `json:"icon,omitempty"`     // 群头像
	Announcement     string            `json:"announcement,omitempty"` // 群公告
	Intro            string            `json:"intro,omitempty"`     // 群简介
	InviteAccountIDs []string          `json:"invite_account_ids"` // 邀请账号ID列表
	InviteMsg        string            `json:"invite_msg"`          // 邀请信息
	Configuration    TeamConfiguration `json:"configuration"`      // 群组配置
}

// TeamConfiguration 群组配置项
type TeamConfiguration struct {
	JoinMode           int `json:"join_mode"`            // 入群模式.默认为0,暂不支持调整. 0（默认）：无需验证，直接入群。 1：需要群主或管理员验证通过才能入群。 2：不允许任何人申请入群。
	InviteMode         int `json:"invite_mode"`          // 邀请权限，即谁可以邀请他人入群。 0（默认）：群主和管理员。 1：所有人。
	UpdateTeamInfoMode int `json:"update_team_info_mode"` // 客户端修改群组信息的权限，即谁可以修改群组信息。 0（默认）：群主和管理员。 1：所有人。
	AgreeMode          int `json:"agree_mode"`          // 邀请入群时是否需要被邀请人的同意。默认为1,暂不支持调整. 0（默认）：需要被邀请人同意才能入群。 1：不需要被邀请人同意，直接入群。
}

// TeamInfoRsp 群组信息响应
type TeamInfoRsp struct {
	TeamID         string            `json:"team_id"`          // 群组ID
	OwnerAccountID string            `json:"owner_account_id"` // 群主账号ID
	Name           string            `json:"name"`             // 群名称
	Icon           string            `json:"icon"`             // 群头像
	Announcement   string            `json:"announcement"`     // 群公告
	Intro          string            `json:"intro"`            // 群简介
	CreateTime     int64             `json:"create_time"`      // 创建时间
	UpdateTime     int64             `json:"update_time"`      // 更新时间
	Configuration  TeamConfiguration `json:"configuration"`   // 群组配置
}

// QueryTeamReq 查询群请求
type QueryTeamReq struct {
	TeamIds  string `json:"team_ids" form:"team_ids"`   // 群组ID列表，多个ID用逗号分隔
	TeamType int    `json:"team_type" form:"team_type"` // 群组类型：1-高级群，2-超大群
}

// TeamInfoV2 群组信息 (v2 API)
type TeamInfoV2 struct {
	TeamID         string `json:"team_id"`          // 群组ID
	Name           string `json:"name"`             // 群名称
	OwnerAccountID string `json:"owner_account_id"`  // 群主账号ID
	CreateTime     int64  `json:"create_time"`     // 创建时间
	TeamType       int    `json:"team_type"`        // 群组类型：1-高级群，2-超大群
	MemberCount    int    `json:"member_count"`     // 成员数量
	MembersLimit   int    `json:"members_limit"`    // 成员上限
}

// ListTeamsData 批量查询群组信息数据体
type ListTeamsData struct {
	InvalidTIDs  []string     `json:"invalid_tids"`   // 无效的群组ID列表
	TeamInfoList []TeamInfoV2 `json:"team_info_list"` // 群组信息列表
}

// UpdateTeamReq 更新群组信息请求
type UpdateTeamReq struct {
	TeamID        string               `json:"team_id"`         // 群组ID
	TeamType      int                  `json:"team_type"`       // 群组类型：1-高级群，2-超大群
	Name          string               `json:"name"`             // 群组名称
	Icon          string               `json:"icon,omitempty"`   // 群组头像URL
	Announcement  string               `json:"announcement,omitempty"` // 群组公告
	Intro         string               `json:"intro,omitempty"`  // 群组简介
	MembersLimit  int                  `json:"members_limit,omitempty"` // 群组成员数上限
	Configuration *TeamConfigurationV3 `json:"configuration,omitempty"` // 群组配置项
	OperatorID    string               `json:"operator_id"`     // 操作者账号ID
}

// TeamConfigurationV3 群组配置项 (v3 API)
type TeamConfigurationV3 struct {
	JoinMode           int `json:"join_mode"`            // 入群模式：0-无需验证，1-需要验证，2-不允许申请
	AgreeMode          int `json:"agree_mode"`           // 同意模式：0-需要被邀请人同意，1-不需要同意
	InviteMode         int `json:"invite_mode"`          // 邀请模式：0-群主和管理员，1-所有人
	UpdateTeamInfoMode int `json:"update_team_info_mode"` // 修改群信息权限：0-群主和管理员，1-所有人
	ChatBannedMode     int `json:"chat_banned_mode"`     // 禁言模式：0-取消禁言，1-全员禁言(除群主管理员)，3-全员禁言
}

// UpdateTeamResp 更新群组信息响应
type UpdateTeamResp struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
	Data struct {
		TeamInfo TeamInfoV3 `json:"team_info"` // 群组信息
	} `json:"data"`
}

// TeamInfoV3 群组信息 (v3 API)
type TeamInfoV3 struct {
	TeamID            string              `json:"team_id"`            // 群组ID
	OwnerAccountID    string              `json:"owner_account_id"`   // 群主账号ID
	Name              string              `json:"name"`               // 群组名称
	Icon              string              `json:"icon"`               // 群组头像URL
	Announcement      string              `json:"announcement"`       // 群组公告
	Intro             string              `json:"intro"`              // 群组简介
	ServerExtension   string              `json:"server_extention"`   // 自定义群组扩展字段
	CustomerExtension string              `json:"customer_extension"` // 客户端自定义扩展字段
	CreateTime        int64               `json:"create_time"`        // 群组创建时间戳
	UpdateTime        int64               `json:"update_time"`        // 群组更新时间戳
	Configuration     TeamConfigurationV3 `json:"configuration"`      // 群组配置项
}

// AddTeamMembersReq 拉人入群请求
type AddTeamMembersReq struct {
	OperatorID       string   `json:"operator_id"`        // 操作人ID
	TeamID           string   `json:"team_id"`            // 群组ID
	TeamType         int      `json:"team_type"`          // 群组类型：1-高级群，2-超大群
	InviteAccountIDs []string `json:"invite_account_ids"` // 邀请账号ID列表
	Msg              string   `json:"msg"`                // 邀请信息
}

// AddTeamMembersFailedItem 拉人入群失败项
type AddTeamMembersFailedItem struct {
	AccountID string `json:"account_id"` // 账号ID
	ErrorCode int    `json:"error_code"`  // 错误码
	ErrorMsg  string `json:"error_msg"`   // 错误信息
}

// AddTeamMembersResp 拉人入群响应
type AddTeamMembersResp struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
	Data struct {
		SuccessList []string                   `json:"success_list"` // 成功账号ID列表
		FailedList  []AddTeamMembersFailedItem `json:"failed_list"`  // 失败项列表
	} `json:"data"`
}

// KickTeamMembersReq 踢人出群请求
type KickTeamMembersReq struct {
	OperatorID     string   `json:"operator_id"`      // 操作人ID
	TeamID         string   `json:"team_id"`          // 群组ID
	TeamType       int      `json:"team_type"`        // 群组类型：1-高级群，2-超大群
	KickAccountIDs []string `json:"kick_account_ids"`  // 踢出账号ID列表，最多10个，最少1个
}

// KickTeamMembersFailedItem 踢人出群失败项
type KickTeamMembersFailedItem struct {
	AccountID string `json:"account_id"` // 账号ID
	ErrorCode int    `json:"error_code"`  // 错误码
	ErrorMsg  string `json:"error_msg"`   // 错误信息
}

// KickTeamMembersResp 踢人出群响应
type KickTeamMembersResp struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
	Data struct {
		SuccessList []string                    `json:"success_list"` // 成功账号ID列表
		FailedList  []KickTeamMembersFailedItem `json:"failed_list"`   // 失败项列表
	} `json:"data"`
}

// DeleteTeamReq 解散群组请求
type DeleteTeamReq struct {
	TeamID     string `json:"team_id"`      // 群组ID
	TeamType   int    `json:"team_type"`    // 群组类型：1-高级群，2-超大群
	OperatorID string `json:"operator_id"`  // 操作人ID
}

// QueryTeamDetailReq 查询群详情请求
type QueryTeamDetailReq struct {
	TeamID   string `json:"team_id" form:"team_id"`     // 群组ID
	TeamType int    `json:"team_type" form:"team_type"` // 群组类型：1-高级群，2-超大群
}

// TeamDetailInfo 群详情信息
type TeamDetailInfo struct {
	TeamID         string `json:"team_id"`          // 群组ID
	TeamType       int    `json:"team_type"`       // 群组类型
	OwnerAccountID string `json:"owner_account_id"` // 群主账号ID
	MemberCount    int    `json:"member_count"`    // 成员数量
	Name           string `json:"name"`            // 群名称
	Announcement   string `json:"announcement"`    // 群公告
	Introduction   string `json:"intro"`          // 群简介
	MembersLimit   int    `json:"members_limit"`   // 成员上限
	CreateTime     int64  `json:"create_time"`     // 创建时间
}

// QueryTeamDetailResp 查询群详情响应
type QueryTeamDetailResp struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
	Data struct {
		TeamInfo TeamDetailInfo `json:"team_info"` // 群详情信息
	} `json:"data"`
}

// LeaveTeamReq 主动退群请求
type LeaveTeamReq struct {
	AccountID string `json:"account_id"` // 账号ID
	TeamID    string `json:"team_id"`     // 群组ID
	TeamType  int    `json:"team_type"`   // 群组类型：1-高级群，2-超大群
}

// GetJoinedTeamsPaginatedReq 分页查询指定账号已加入的群组信息请求
type GetJoinedTeamsPaginatedReq struct {
	AccountID string `form:"account_id"` // 账号ID
	TeamType  int    `form:"team_type"`   // 群组类型：1-高级群，2-超大群
	PageToken string `form:"page_token"`  // 分页token
	Limit     int    `form:"limit"`       // 分页大小
}

// TeamInfo 群组信息
type TeamInfo struct {
	Name          string             `json:"name"`          // 群名称
	Configuration *TeamConfiguration `json:"configuration"` // 群组配置
	MemberCount   int                `json:"member_count"`   // 成员数量
	TeamID        string             `json:"team_id"`        // 群组ID
	CreateTime    int64              `json:"create_time"`   // 创建时间
}

// GetJoinedTeamsPaginatedResp 分页查询指定账号已加入的群组信息响应
type GetJoinedTeamsPaginatedResp struct {
	PageNext     string     `json:"page_next"`      // 下一页token
	HasMore      bool       `json:"has_more"`       // 是否还有更多
	TeamInfoList []TeamInfo `json:"team_info_list"` // 群组信息列表
}

// MuteAllReq 全员禁言/解除禁言请求
type MuteAllReq struct {
	TeamID         string `json:"team_id"`          // 群组ID
	TeamType       int    `json:"team_type"`        // 群组类型：1-高级群，2-超大群
	OperatorID     string `json:"operator_id"`       // 操作人ID
	ChatBannedMode int    `json:"chat_banned_mode"` // 禁言模式. 0（默认）：取消群组禁言。1：禁言全体普通成员，不包括群主和管理员。3：禁言全体成员。
}

// ListTeamMembersReq 分页查询群成员列表请求
type ListTeamMembersReq struct {
	TeamID     string `json:"team_id" form:"team_id"`     // 群组ID
	TeamType   int    `json:"team_type" form:"team_type"` // 群组类型：1-高级群，2-超大群
	Descending bool   `json:"descending" form:"descending"` // 是否按照成员入群时间降序排列
	PageToken  string `json:"page_token" form:"page_token"` // 分页token
	Limit      int    `json:"limit" form:"limit"`           // 每页返回的成员数上限，最大为 100
}

// ListTeamMemberItem 群成员条目
type ListTeamMemberItem struct {
	MemberRole int    `json:"member_role"` // 成员角色,群成员类型。0：普通成员。1：群主。2：群管理员。
	AccountID  string `json:"account_id"`  // 账号ID
	TeamNick   string `json:"team_nick"`    // 群昵称
	JoinTime   int64  `json:"join_time"`   // 加入时间
}

// ListTeamMembersData 数据体
type ListTeamMembersData struct {
	HasMore   bool                 `json:"has_more"`   // 是否还有更多
	NextToken string               `json:"next_token"` // 下一页token
	Items     []ListTeamMemberItem `json:"items"`      // 成员列表
}

// ListTeamMembersResp 分页查询群成员列表响应
type ListTeamMembersResp struct {
	Code int                 `json:"code"` // 状态码
	Msg  string              `json:"msg"`   // 提示信息
	Data ListTeamMembersData `json:"data"` // 数据体
}

// ========== 消息管理相关类型 ==========

// MessageInfo 消息信息
type MessageInfo struct {
	FromUID     string     `json:"from_uid"`     // 会话消息发送者的账号 ID
	ToUID       string     `json:"to_uid"`        // 接收者账号 ID
	ToType      string     `json:"to_type"`      // 会话类型，1：单聊会话；2：高级群会话；
	MessageType int        `json:"message_type"` // 消息类型。 0：文本消息 1：图片消息 2：语音消息 3：视频消息 4：地理位置消息 6：文件消息
	Text        string     `json:"text"`         // 对于文本消息和提示消息，该字段必填，值为消息内容，长度上限 5000 位字符。
	FromBuss    string     `json:"from_buss"`    // 来源业务
	Attachment  Attachment `json:"attachment"`   // 附件
}

// Attachment 附件
type Attachment struct {
	Name     string  `json:"name,omitempty"`     // 文件/图片/音频/视频/位置名称
	Md5      string  `json:"md5,omitempty"`      // 文件/图片/音频/视频md5
	Url      string  `json:"url,omitempty"`      // 文件/图片/音频/视频URL
	Ext      string  `json:"ext,omitempty"`      // 文件/图片/音频/视频后缀/格式
	Size     int     `json:"size,omitempty"`     // 文件/图片/音频/视频大小（字节）
	Width    int     `json:"w,omitempty"`        // 图片/视频宽度（像素）
	Height   int     `json:"h,omitempty"`        // 图片/视频高度（像素）
	Duration int     `json:"dur,omitempty"`      // 音频/视频持续时长（ms）
	Lng      float64 `json:"lng,omitempty"`      // 经度
	Lat      float64 `json:"lat,omitempty"`      // 纬度
}

// BroadcastMessageInfo 广播消息信息
type BroadcastMessageInfo struct {
	Content       string   `json:"content"`        // 广播消息内容，长度上限 4096 个字符
	FromAccountID string   `json:"from_account_id"` // 广播消息发送者的云信账号 ID
	TTL           int      `json:"ttl"`            // 存离线的有效时间，单位为小时，默认为 168 小时，即 7 天
	TargetOS      []string `json:"target_os,omitempty"` // 接收广播消息的目标客户端，默认为所有客户端
	ResendFlag    int      `json:"resend_flag,omitempty"` // 是否为重发消息。0（默认）：非重发消息；1：重发消息
}

// BroadcastNotificationResp 广播消息响应
type BroadcastNotificationResp struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
	Data struct {
		BroadcastID       string `json:"broadcast_id"`        // 广播消息的 ID
		CreateTime        int64  `json:"create_time"`        // 广播消息发送时间戳
		ExpireTime        int64  `json:"expire_time"`         // 广播消息过期时间戳
		ClientBroadcastID string `json:"client_broadcast_id"` // 聊天室全服广播消息 ID
		SenderID          string `json:"sender_id"`           // 聊天室广播消息发送者的云信账号 ID
		Extension         string `json:"extension"`            // 开发者扩展字段
		ResendFlag        int    `json:"resend_flag"`          // 是否为重发消息
	} `json:"data"`
}

// ========== 认证相关类型 ==========

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
}

// LoginReply 登录响应
type LoginReply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Token    string `json:"token"`
		UserKey  string `json:"userKey,omitempty"`
	} `json:"data"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Email    string `json:"email,omitempty"` // 邮箱
}

// RegisterReply 注册响应
type RegisterReply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

// LogoutReply 登出响应
type LogoutReply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

// ========== 元信息相关类型 ==========

// ProviderStatusResp 返回当前提供商及启用渠道
type ProviderStatusResp struct {
	Code             int      `json:"code"`              // 状态码
	Msg              string   `json:"msg"`              // 提示信息
	CurrentProvider  string   `json:"current_provider"`  // 当前使用的IM提供商
	EnabledProviders []string `json:"enabled_providers"` // 已启用的IM提供商列表
}

