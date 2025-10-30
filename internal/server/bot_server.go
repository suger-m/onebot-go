package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	types "onebot-go2/pkg/const"
	"onebot-go2/pkg/event"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	token        string
	activeClient *websocket.Conn  // 改为单客户端连接
	clientMu     sync.RWMutex     // 保护客户端连接的读写锁
	upgrader     websocket.Upgrader
	pendingCalls sync.Map         // 存储待响应的 API 调用
	dispatcher   *event.Dispatcher
	echoCounter  uint64           // Echo ID 计数器
	callTimeout  time.Duration    // API 调用超时时间
	connected    atomic.Bool      // 连接状态
}

func NewWSServer(token string) *WSServer {
	server := &WSServer{
		token:       token,
		dispatcher:  event.NewDispatcher(),
		callTimeout: 10 * time.Second, // 默认10秒超时
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	server.connected.Store(false)
	return server
}

// GetDispatcher 获取事件分发器
func (s *WSServer) GetDispatcher() *event.Dispatcher {
	return s.dispatcher
}

// SetCallTimeout 设置 API 调用超时时间
func (s *WSServer) SetCallTimeout(timeout time.Duration) {
	s.callTimeout = timeout
}

// IsConnected 检查是否已连接
func (s *WSServer) IsConnected() bool {
	return s.connected.Load()
}

// generateEcho 生成唯一的 echo ID
func (s *WSServer) generateEcho() string {
	// 使用 UUID 确保唯一性
	return uuid.New().String()
}

func (s *WSServer) HandlerWebsocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrader error: %v", err)
		return
	}

	// 设置活动客户端连接
	s.clientMu.Lock()
	if s.activeClient != nil {
		// 关闭旧连接
		s.activeClient.Close()
		log.Printf("Closed old connection")
	}
	s.activeClient = conn
	s.connected.Store(true)
	s.clientMu.Unlock()

	defer func() {
		s.clientMu.Lock()
		if s.activeClient == conn {
			s.activeClient = nil
			s.connected.Store(false)
		}
		s.clientMu.Unlock()
		conn.Close()
	}()

	log.Printf("WebSocket connection established from %s", conn.RemoteAddr())

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", conn.RemoteAddr(), err)
			break
		}

		// 尝试解析为 API 响应
		var response struct {
			Echo string `json:"echo"`
			types.APIResponse
		}
		if err := json.Unmarshal(message, &response); err == nil && response.Echo != "" {
			// 这是一个 API 响应
			if ch, ok := s.pendingCalls.LoadAndDelete(response.Echo); ok {
				if respChan, ok := ch.(chan *types.APIResponse); ok {
					select {
					case respChan <- &response.APIResponse:
					case <-time.After(1 * time.Second):
						log.Printf("Response channel timeout for echo: %s", response.Echo)
					}
					close(respChan)
				}
			}
			continue
		}

		// 解析为事件
		evt, err := ParseEvent(message)
		if err != nil {
			log.Printf("Parse event error: %v", err)
			continue
		}

		// 分发事件到注册的处理器
		if err := s.dispatcher.Dispatch(context.Background(), evt, s); err != nil {
			log.Printf("Error dispatching event: %v", err)
		}
	}
}

func ParseEvent(data []byte) (interface{}, error) {
	var base types.Event
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.PostType {
	case types.PostTypeMessage:
		var event types.MessageEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	case types.PostTypeNotice:
		var event types.NoticeEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	case types.PostTypeRequest:
		var event types.RequestEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	case types.PostTypeMetaEvent:
		var event types.MetaEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil

	default:
		return nil, fmt.Errorf("unknown post_type: %v", base.PostType)
	}
}

// CallAPI 通用 API 调用方法
func (s *WSServer) CallAPI(action string, params interface{}) (*types.APIResponse, error) {
	if !s.IsConnected() {
		return nil, errors.New("not connected to OneBot client")
	}

	// 生成唯一的 echo ID
	echo := s.generateEcho()

	// 创建响应通道
	respChan := make(chan *types.APIResponse, 1)
	s.pendingCalls.Store(echo, respChan)

	// 构造 API 请求
	request := types.APIRequest{
		Action: action,
		Params: params,
		Echo:   echo,
	}

	// 序列化请求
	data, err := json.Marshal(request)
	if err != nil {
		s.pendingCalls.Delete(echo)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	s.clientMu.RLock()
	conn := s.activeClient
	s.clientMu.RUnlock()

	if conn == nil {
		s.pendingCalls.Delete(echo)
		return nil, errors.New("no active connection")
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.pendingCalls.Delete(echo)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 等待响应（带超时）
	select {
	case resp := <-respChan:
		if resp.Status != "ok" && resp.Status != "async" {
			return resp, fmt.Errorf("API call failed: %s (retcode: %d)", resp.Message, resp.RetCode)
		}
		return resp, nil
	case <-time.After(s.callTimeout):
		s.pendingCalls.Delete(echo)
		return nil, errors.New("API call timeout")
	}
}

// ============ 消息相关 API ============

// SendPrivateMsg 发送私聊消息
func (s *WSServer) SendPrivateMsg(userID int64, message types.MessageArray) (*types.SendMessageResponse, error) {
	params := types.SendMessageParams{
		MessageType: types.MessageTypePrivate,
		UserID:      userID,
		Message:     message,
	}

	resp, err := s.CallAPI(types.ActionSendPrivateMsg, params)
	if err != nil {
		return nil, err
	}

	var result types.SendMessageResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// SendGroupMsg 发送群消息
func (s *WSServer) SendGroupMsg(groupID int64, message types.MessageArray) (*types.SendMessageResponse, error) {
	params := types.SendMessageParams{
		MessageType: types.MessageTypeGroup,
		GroupID:     groupID,
		Message:     message,
	}

	resp, err := s.CallAPI(types.ActionSendGroupMsg, params)
	if err != nil {
		return nil, err
	}

	var result types.SendMessageResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// SendMsg 发送消息（自动识别类型）
func (s *WSServer) SendMsg(params *types.SendMessageParams) (*types.SendMessageResponse, error) {
	resp, err := s.CallAPI(types.ActionSendMsg, params)
	if err != nil {
		return nil, err
	}

	var result types.SendMessageResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DeleteMsg 撤回消息
func (s *WSServer) DeleteMsg(messageID int32) error {
	params := types.DeleteMsgParams{
		MessageID: messageID,
	}

	_, err := s.CallAPI(types.ActionDeleteMsg, params)
	return err
}

// GetMsg 获取消息
func (s *WSServer) GetMsg(messageID int32) (*types.GetMsgResponse, error) {
	params := types.GetMsgParams{
		MessageID: messageID,
	}

	resp, err := s.CallAPI(types.ActionGetMsg, params)
	if err != nil {
		return nil, err
	}

	var result types.GetMsgResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetForwardMsg 获取合并转发消息
func (s *WSServer) GetForwardMsg(id string) (*types.GetForwardMsgResponse, error) {
	params := types.GetForwardMsgParams{
		ID: id,
	}

	resp, err := s.CallAPI(types.ActionGetForwardMsg, params)
	if err != nil {
		return nil, err
	}

	var result types.GetForwardMsgResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// SendLike 发送好友赞
func (s *WSServer) SendLike(userID int64, times int) error {
	params := types.SendLikeParams{
		UserID: userID,
		Times:  times,
	}

	_, err := s.CallAPI(types.ActionSendLike, params)
	return err
}

// ============ 群管理相关 API ============

// SetGroupKick 群组踢人
func (s *WSServer) SetGroupKick(groupID, userID int64, rejectAddRequest bool) error {
	params := types.SetGroupKickParams{
		GroupID:          groupID,
		UserID:           userID,
		RejectAddRequest: rejectAddRequest,
	}

	_, err := s.CallAPI(types.ActionSetGroupKick, params)
	return err
}

// SetGroupBan 群组单人禁言
func (s *WSServer) SetGroupBan(groupID, userID int64, duration int64) error {
	params := types.SetGroupBanParams{
		GroupID:  groupID,
		UserID:   userID,
		Duration: duration,
	}

	_, err := s.CallAPI(types.ActionSetGroupBan, params)
	return err
}

// SetGroupAnonymousBan 群组匿名用户禁言
func (s *WSServer) SetGroupAnonymousBan(groupID int64, flag string, duration int64) error {
	params := types.SetGroupAnonymousBanParams{
		GroupID:  groupID,
		Flag:     flag,
		Duration: duration,
	}

	_, err := s.CallAPI(types.ActionSetGroupAnonymousBan, params)
	return err
}

// SetGroupWholeBan 群组全员禁言
func (s *WSServer) SetGroupWholeBan(groupID int64, enable bool) error {
	params := types.SetGroupWholeBanParams{
		GroupID: groupID,
		Enable:  enable,
	}

	_, err := s.CallAPI(types.ActionSetGroupWholeBan, params)
	return err
}

// SetGroupAdmin 设置群管理员
func (s *WSServer) SetGroupAdmin(groupID, userID int64, enable bool) error {
	params := types.SetGroupAdminParams{
		GroupID: groupID,
		UserID:  userID,
		Enable:  enable,
	}

	_, err := s.CallAPI(types.ActionSetGroupAdmin, params)
	return err
}

// SetGroupAnonymous 设置群匿名
func (s *WSServer) SetGroupAnonymous(groupID int64, enable bool) error {
	params := types.SetGroupAnonymousParams{
		GroupID: groupID,
		Enable:  enable,
	}

	_, err := s.CallAPI(types.ActionSetGroupAnonymous, params)
	return err
}

// SetGroupCard 设置群名片
func (s *WSServer) SetGroupCard(groupID, userID int64, card string) error {
	params := types.SetGroupCardParams{
		GroupID: groupID,
		UserID:  userID,
		Card:    card,
	}

	_, err := s.CallAPI(types.ActionSetGroupCard, params)
	return err
}

// SetGroupName 设置群名
func (s *WSServer) SetGroupName(groupID int64, groupName string) error {
	params := types.SetGroupNameParams{
		GroupID:   groupID,
		GroupName: groupName,
	}

	_, err := s.CallAPI(types.ActionSetGroupName, params)
	return err
}

// SetGroupLeave 退出群组
func (s *WSServer) SetGroupLeave(groupID int64, isDismiss bool) error {
	params := types.SetGroupLeaveParams{
		GroupID:   groupID,
		IsDismiss: isDismiss,
	}

	_, err := s.CallAPI(types.ActionSetGroupLeave, params)
	return err
}

// SetGroupSpecialTitle 设置群组专属头衔
func (s *WSServer) SetGroupSpecialTitle(groupID, userID int64, specialTitle string, duration int64) error {
	params := types.SetGroupSpecialTitleParams{
		GroupID:      groupID,
		UserID:       userID,
		SpecialTitle: specialTitle,
		Duration:     duration,
	}

	_, err := s.CallAPI(types.ActionSetGroupSpecialTitle, params)
	return err
}

// ============ 请求处理相关 API ============

// SetFriendAddRequest 处理加好友请求
func (s *WSServer) SetFriendAddRequest(flag string, approve bool, remark string) error {
	params := types.SetFriendAddRequestParams{
		Flag:    flag,
		Approve: approve,
		Remark:  remark,
	}

	_, err := s.CallAPI(types.ActionSetFriendAddRequest, params)
	return err
}

// SetGroupAddRequest 处理加群请求/邀请
func (s *WSServer) SetGroupAddRequest(flag, subType string, approve bool, reason string) error {
	params := types.SetGroupAddRequestParams{
		Flag:    flag,
		SubType: subType,
		Approve: approve,
		Reason:  reason,
	}

	_, err := s.CallAPI(types.ActionSetGroupAddRequest, params)
	return err
}

// ============ 信息获取相关 API ============

// GetLoginInfo 获取登录号信息
func (s *WSServer) GetLoginInfo() (*types.GetLoginInfoResponse, error) {
	resp, err := s.CallAPI(types.ActionGetLoginInfo, nil)
	if err != nil {
		return nil, err
	}

	var result types.GetLoginInfoResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetStrangerInfo 获取陌生人信息
func (s *WSServer) GetStrangerInfo(userID int64, noCache bool) (*types.GetStrangerInfoResponse, error) {
	params := types.GetStrangerInfoParams{
		UserID:  userID,
		NoCache: noCache,
	}

	resp, err := s.CallAPI(types.ActionGetStrangerInfo, params)
	if err != nil {
		return nil, err
	}

	var result types.GetStrangerInfoResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetFriendList 获取好友列表
func (s *WSServer) GetFriendList() (types.GetFriendListResponse, error) {
	resp, err := s.CallAPI(types.ActionGetFriendList, nil)
	if err != nil {
		return nil, err
	}

	var result types.GetFriendListResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// GetGroupInfo 获取群信息
func (s *WSServer) GetGroupInfo(groupID int64, noCache bool) (*types.GetGroupInfoResponse, error) {
	params := types.GetGroupInfoParams{
		GroupID: groupID,
		NoCache: noCache,
	}

	resp, err := s.CallAPI(types.ActionGetGroupInfo, params)
	if err != nil {
		return nil, err
	}

	var result types.GetGroupInfoResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetGroupList 获取群列表
func (s *WSServer) GetGroupList() (types.GetGroupListResponse, error) {
	resp, err := s.CallAPI(types.ActionGetGroupList, nil)
	if err != nil {
		return nil, err
	}

	var result types.GetGroupListResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// GetGroupMemberInfo 获取群成员信息
func (s *WSServer) GetGroupMemberInfo(groupID, userID int64, noCache bool) (*types.GetGroupMemberInfoResponse, error) {
	params := types.GetGroupMemberInfoParams{
		GroupID: groupID,
		UserID:  userID,
		NoCache: noCache,
	}

	resp, err := s.CallAPI(types.ActionGetGroupMemberInfo, params)
	if err != nil {
		return nil, err
	}

	var result types.GetGroupMemberInfoResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetGroupMemberList 获取群成员列表
func (s *WSServer) GetGroupMemberList(groupID int64) (types.GetGroupMemberListResponse, error) {
	params := types.GetGroupMemberListParams{
		GroupID: groupID,
	}

	resp, err := s.CallAPI(types.ActionGetGroupMemberList, params)
	if err != nil {
		return nil, err
	}

	var result types.GetGroupMemberListResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// GetGroupHonorInfo 获取群荣誉信息
func (s *WSServer) GetGroupHonorInfo(groupID int64, honorType string) (*types.GetGroupHonorInfoResponse, error) {
	params := types.GetGroupHonorInfoParams{
		GroupID: groupID,
		Type:    honorType,
	}

	resp, err := s.CallAPI(types.ActionGetGroupHonorInfo, params)
	if err != nil {
		return nil, err
	}

	var result types.GetGroupHonorInfoResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetCookies 获取Cookies
func (s *WSServer) GetCookies(domain string) (*types.GetCookiesResponse, error) {
	params := types.GetCookiesParams{
		Domain: domain,
	}

	resp, err := s.CallAPI(types.ActionGetCookies, params)
	if err != nil {
		return nil, err
	}

	var result types.GetCookiesResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetCsrfToken 获取CSRF Token
func (s *WSServer) GetCsrfToken() (*types.GetCsrfTokenResponse, error) {
	resp, err := s.CallAPI(types.ActionGetCsrfToken, nil)
	if err != nil {
		return nil, err
	}

	var result types.GetCsrfTokenResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetCredentials 获取QQ相关接口凭证
func (s *WSServer) GetCredentials(domain string) (*types.GetCredentialsResponse, error) {
	params := types.GetCredentialsParams{
		Domain: domain,
	}

	resp, err := s.CallAPI(types.ActionGetCredentials, params)
	if err != nil {
		return nil, err
	}

	var result types.GetCredentialsResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetStatus 获取运行状态
func (s *WSServer) GetStatus() (*types.GetStatusResponse, error) {
	resp, err := s.CallAPI(types.ActionGetStatus, nil)
	if err != nil {
		return nil, err
	}

	var result types.GetStatusResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetVersionInfo 获取版本信息
func (s *WSServer) GetVersionInfo() (*types.GetVersionInfoResponse, error) {
	resp, err := s.CallAPI(types.ActionGetVersionInfo, nil)
	if err != nil {
		return nil, err
	}

	var result types.GetVersionInfoResponse
	if err := mapToStruct(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ============ 辅助函数 ============

// mapToStruct 将 map 转换为结构体
func mapToStruct(data interface{}, result interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, result)
}
