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
