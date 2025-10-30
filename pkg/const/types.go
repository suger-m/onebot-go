package types

import "encoding/json"

// MessageType 消息类型
type MessageType string

const (
	MessageTypePrivate MessageType = "private"
	MessageTypeGroup   MessageType = "group"
)

// PostType 上报类型
type PostType string

const (
	PostTypeMessage   PostType = "message"
	PostTypeNotice    PostType = "notice"
	PostTypeRequest   PostType = "request"
	PostTypeMetaEvent PostType = "meta_event"
)

// Sender 发送者信息
type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex,omitempty"`
	Age      int    `json:"age,omitempty"`
	Card     string `json:"card,omitempty"`
	Area     string `json:"area,omitempty"`
	Level    string `json:"level,omitempty"`
	Role     string `json:"role,omitempty"`
	Title    string `json:"title,omitempty"`
}

// Message 消息段
type Message struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// MessageArray 消息数组
type MessageArray []Message

// Event 事件基类
type Event struct {
	Time       int64           `json:"time"`
	SelfID     int64           `json:"self_id"`
	PostType   PostType        `json:"post_type"`
	RawMessage json.RawMessage `json:"-"`
}

// MessageEvent 消息事件
type MessageEvent struct {
	Event
	MessageType MessageType  `json:"message_type"`
	SubType     string       `json:"sub_type"`
	MessageID   int32        `json:"message_id"`
	UserID      int64        `json:"user_id"`
	Message     MessageArray `json:"message"`
	RawMessage  string       `json:"raw_message"`
	Font        int32        `json:"font"`
	Sender      Sender       `json:"sender"`
	GroupID     int64        `json:"group_id,omitempty"`
}

// NoticeEvent 通知事件
type NoticeEvent struct {
	Event
	NoticeType string `json:"notice_type"`
	SubType    string `json:"sub_type,omitempty"`
	UserID     int64  `json:"user_id"`
	GroupID    int64  `json:"group_id,omitempty"`
	OperatorID int64  `json:"operator_id,omitempty"`
}

// RequestEvent 请求事件
type RequestEvent struct {
	Event
	RequestType string `json:"request_type"`
	SubType     string `json:"sub_type,omitempty"`
	UserID      int64  `json:"user_id"`
	GroupID     int64  `json:"group_id,omitempty"`
	Comment     string `json:"comment"`
	Flag        string `json:"flag"`
}

// MetaEvent 元事件
type MetaEvent struct {
	Event
	MetaEventType string                 `json:"meta_event_type"`
	SubType       string                 `json:"sub_type,omitempty"`
	Status        map[string]interface{} `json:"status,omitempty"`
	Interval      int64                  `json:"interval,omitempty"`
}

// APIResponse API响应
type APIResponse struct {
	Status  string      `json:"status"`
	RetCode int         `json:"retcode"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Wording string      `json:"wording,omitempty"`
}

// SendMessageParams 发送消息参数
type SendMessageParams struct {
	MessageType MessageType  `json:"message_type,omitempty"`
	UserID      int64        `json:"user_id,omitempty"`
	GroupID     int64        `json:"group_id,omitempty"`
	Message     MessageArray `json:"message"`
	AutoEscape  bool         `json:"auto_escape,omitempty"`
}

// SendMessageResponse 发送消息响应
type SendMessageResponse struct {
	MessageID int32 `json:"message_id"`
}

// DeleteMsgParams 撤回消息参数
type DeleteMsgParams struct {
	MessageID int32 `json:"message_id"`
}

// GetMsgParams 获取消息参数
type GetMsgParams struct {
	MessageID int32 `json:"message_id"`
}

// GetMsgResponse 获取消息响应
type GetMsgResponse struct {
	Time        int64        `json:"time"`
	MessageType MessageType  `json:"message_type"`
	MessageID   int32        `json:"message_id"`
	RealID      int32        `json:"real_id"`
	Sender      Sender       `json:"sender"`
	Message     MessageArray `json:"message"`
}

// GetForwardMsgParams 获取合并转发消息参数
type GetForwardMsgParams struct {
	ID string `json:"id"`
}

// GetForwardMsgResponse 获取合并转发消息响应
type GetForwardMsgResponse struct {
	Message MessageArray `json:"message"`
}

// SendLikeParams 点赞参数
type SendLikeParams struct {
	UserID int64 `json:"user_id"`
	Times  int   `json:"times"`
}

// SetGroupKickParams 群组踢人参数
type SetGroupKickParams struct {
	GroupID          int64 `json:"group_id"`
	UserID           int64 `json:"user_id"`
	RejectAddRequest bool  `json:"reject_add_request,omitempty"`
}

// SetGroupBanParams 群组单人禁言参数
type SetGroupBanParams struct {
	GroupID  int64 `json:"group_id"`
	UserID   int64 `json:"user_id"`
	Duration int64 `json:"duration"` // 禁言时长，单位秒，0表示取消禁言
}

// SetGroupAnonymousBanParams 群组匿名用户禁言参数
type SetGroupAnonymousBanParams struct {
	GroupID   int64  `json:"group_id"`
	Anonymous string `json:"anonymous,omitempty"`
	Flag      string `json:"flag,omitempty"`
	Duration  int64  `json:"duration"`
}

// SetGroupWholeBanParams 群组全员禁言参数
type SetGroupWholeBanParams struct {
	GroupID int64 `json:"group_id"`
	Enable  bool  `json:"enable"`
}

// SetGroupAdminParams 设置群管理员参数
type SetGroupAdminParams struct {
	GroupID int64 `json:"group_id"`
	UserID  int64 `json:"user_id"`
	Enable  bool  `json:"enable"`
}

// SetGroupAnonymousParams 设置群匿名参数
type SetGroupAnonymousParams struct {
	GroupID int64 `json:"group_id"`
	Enable  bool  `json:"enable"`
}

// SetGroupCardParams 设置群名片参数
type SetGroupCardParams struct {
	GroupID int64  `json:"group_id"`
	UserID  int64  `json:"user_id"`
	Card    string `json:"card"`
}

// SetGroupNameParams 设置群名参数
type SetGroupNameParams struct {
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
}

// SetGroupLeaveParams 退出群组参数
type SetGroupLeaveParams struct {
	GroupID   int64 `json:"group_id"`
	IsDismiss bool  `json:"is_dismiss,omitempty"`
}

// SetGroupSpecialTitleParams 设置群组专属头衔参数
type SetGroupSpecialTitleParams struct {
	GroupID      int64  `json:"group_id"`
	UserID       int64  `json:"user_id"`
	SpecialTitle string `json:"special_title"`
	Duration     int64  `json:"duration,omitempty"` // 专属头衔有效期，单位秒，-1表示永久
}

// SetFriendAddRequestParams 处理加好友请求参数
type SetFriendAddRequestParams struct {
	Flag    string `json:"flag"`
	Approve bool   `json:"approve"`
	Remark  string `json:"remark,omitempty"`
}

// SetGroupAddRequestParams 处理加群请求/邀请参数
type SetGroupAddRequestParams struct {
	Flag    string `json:"flag"`
	SubType string `json:"sub_type"` // add 或 invite
	Approve bool   `json:"approve"`
	Reason  string `json:"reason,omitempty"`
}

// GetLoginInfoResponse 获取登录号信息响应
type GetLoginInfoResponse struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
}

// GetStrangerInfoParams 获取陌生人信息参数
type GetStrangerInfoParams struct {
	UserID  int64 `json:"user_id"`
	NoCache bool  `json:"no_cache,omitempty"`
}

// GetStrangerInfoResponse 获取陌生人信息响应
type GetStrangerInfoResponse struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
}

// FriendInfo 好友信息
type FriendInfo struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
}

// GetFriendListResponse 获取好友列表响应
type GetFriendListResponse []FriendInfo

// GetGroupInfoParams 获取群信息参数
type GetGroupInfoParams struct {
	GroupID int64 `json:"group_id"`
	NoCache bool  `json:"no_cache,omitempty"`
}

// GroupInfo 群信息
type GroupInfo struct {
	GroupID         int64  `json:"group_id"`
	GroupName       string `json:"group_name"`
	MemberCount     int    `json:"member_count"`
	MaxMemberCount  int    `json:"max_member_count"`
	GroupCreateTime int64  `json:"group_create_time,omitempty"`
	GroupLevel      int    `json:"group_level,omitempty"`
}

// GetGroupInfoResponse 获取群信息响应
type GetGroupInfoResponse GroupInfo

// GetGroupListResponse 获取群列表响应
type GetGroupListResponse []GroupInfo

// GetGroupMemberInfoParams 获取群成员信息参数
type GetGroupMemberInfoParams struct {
	GroupID int64 `json:"group_id"`
	UserID  int64 `json:"user_id"`
	NoCache bool  `json:"no_cache,omitempty"`
}

// GroupMemberInfo 群成员信息
type GroupMemberInfo struct {
	GroupID         int64  `json:"group_id"`
	UserID          int64  `json:"user_id"`
	Nickname        string `json:"nickname"`
	Card            string `json:"card"`
	Sex             string `json:"sex"`
	Age             int    `json:"age"`
	Area            string `json:"area"`
	JoinTime        int64  `json:"join_time"`
	LastSentTime    int64  `json:"last_sent_time"`
	Level           string `json:"level"`
	Role            string `json:"role"`
	Unfriendly      bool   `json:"unfriendly"`
	Title           string `json:"title"`
	TitleExpireTime int64  `json:"title_expire_time"`
	CardChangeable  bool   `json:"card_changeable"`
}

// GetGroupMemberInfoResponse 获取群成员信息响应
type GetGroupMemberInfoResponse GroupMemberInfo

// GetGroupMemberListParams 获取群成员列表参数
type GetGroupMemberListParams struct {
	GroupID int64 `json:"group_id"`
}

// GetGroupMemberListResponse 获取群成员列表响应
type GetGroupMemberListResponse []GroupMemberInfo

// GetGroupHonorInfoParams 获取群荣誉信息参数
type GetGroupHonorInfoParams struct {
	GroupID int64  `json:"group_id"`
	Type    string `json:"type"` // talkative, performer, legend, strong_newbie, emotion, all
}

// HonorInfo 荣誉信息
type HonorInfo struct {
	UserID      int64  `json:"user_id"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
}

// GetGroupHonorInfoResponse 获取群荣誉信息响应
type GetGroupHonorInfoResponse struct {
	GroupID            int64       `json:"group_id"`
	CurrentTalkative   *HonorInfo  `json:"current_talkative,omitempty"`
	TalkativeList      []HonorInfo `json:"talkative_list,omitempty"`
	PerformerList      []HonorInfo `json:"performer_list,omitempty"`
	LegendList         []HonorInfo `json:"legend_list,omitempty"`
	StrongNewbieList   []HonorInfo `json:"strong_newbie_list,omitempty"`
	EmotionList        []HonorInfo `json:"emotion_list,omitempty"`
}

// GetCookiesParams 获取Cookies参数
type GetCookiesParams struct {
	Domain string `json:"domain,omitempty"`
}

// GetCookiesResponse 获取Cookies响应
type GetCookiesResponse struct {
	Cookies string `json:"cookies"`
}

// GetCsrfTokenResponse 获取CSRF Token响应
type GetCsrfTokenResponse struct {
	Token int `json:"token"`
}

// GetCredentialsParams 获取QQ相关接口凭证参数
type GetCredentialsParams struct {
	Domain string `json:"domain,omitempty"`
}

// GetCredentialsResponse 获取QQ相关接口凭证响应
type GetCredentialsResponse struct {
	Cookies string `json:"cookies"`
	Token   int    `json:"csrf_token"`
}

// GetRecordParams 获取语音参数
type GetRecordParams struct {
	File      string `json:"file"`
	OutFormat string `json:"out_format"`
}

// GetRecordResponse 获取语音响应
type GetRecordResponse struct {
	File string `json:"file"`
}

// GetImageParams 获取图片参数
type GetImageParams struct {
	File string `json:"file"`
}

// GetImageResponse 获取图片响应
type GetImageResponse struct {
	File string `json:"file"`
}

// CanSendImageResponse 检查是否可以发送图片响应
type CanSendImageResponse struct {
	Yes bool `json:"yes"`
}

// CanSendRecordResponse 检查是否可以发送语音响应
type CanSendRecordResponse struct {
	Yes bool `json:"yes"`
}

// GetStatusResponse 获取状态响应
type GetStatusResponse struct {
	Online bool                   `json:"online"`
	Good   bool                   `json:"good"`
	Stat   map[string]interface{} `json:"stat,omitempty"`
}

// GetVersionInfoResponse 获取版本信息响应
type GetVersionInfoResponse struct {
	AppName         string `json:"app_name"`
	AppVersion      string `json:"app_version"`
	ProtocolVersion string `json:"protocol_version"`
}

// SetRestartParams 重启参数
type SetRestartParams struct {
	Delay int `json:"delay,omitempty"` // 延迟重启，单位毫秒
}

// CleanCacheParams 清理缓存参数
type CleanCacheParams struct{}

// APIRequest API请求
type APIRequest struct {
	Action string      `json:"action"`
	Params interface{} `json:"params,omitempty"`
	Echo   string      `json:"echo,omitempty"`
}
