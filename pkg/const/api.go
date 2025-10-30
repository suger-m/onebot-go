package types

// OneBot API Action 常量定义

const (
	// 消息相关
	ActionSendPrivateMsg = "send_private_msg" // 发送私聊消息
	ActionSendGroupMsg   = "send_group_msg"   // 发送群消息
	ActionSendMsg        = "send_msg"         // 发送消息
	ActionDeleteMsg      = "delete_msg"       // 撤回消息
	ActionGetMsg         = "get_msg"          // 获取消息
	ActionGetForwardMsg  = "get_forward_msg"  // 获取合并转发消息
	ActionSendLike       = "send_like"        // 发送好友赞

	// 群管理相关
	ActionSetGroupKick         = "set_group_kick"          // 群组踢人
	ActionSetGroupBan          = "set_group_ban"           // 群组单人禁言
	ActionSetGroupAnonymousBan = "set_group_anonymous_ban" // 群组匿名用户禁言
	ActionSetGroupWholeBan     = "set_group_whole_ban"     // 群组全员禁言
	ActionSetGroupAdmin        = "set_group_admin"         // 群组设置管理员
	ActionSetGroupAnonymous    = "set_group_anonymous"     // 群组匿名
	ActionSetGroupCard         = "set_group_card"          // 设置群名片（群备注）
	ActionSetGroupName         = "set_group_name"          // 设置群名
	ActionSetGroupLeave        = "set_group_leave"         // 退出群组
	ActionSetGroupSpecialTitle = "set_group_special_title" // 设置群组专属头衔

	// 请求处理相关
	ActionSetFriendAddRequest = "set_friend_add_request" // 处理加好友请求
	ActionSetGroupAddRequest  = "set_group_add_request"  // 处理加群请求/邀请

	// 信息获取相关
	ActionGetLoginInfo        = "get_login_info"         // 获取登录号信息
	ActionGetStrangerInfo     = "get_stranger_info"      // 获取陌生人信息
	ActionGetFriendList       = "get_friend_list"        // 获取好友列表
	ActionGetGroupInfo        = "get_group_info"         // 获取群信息
	ActionGetGroupList        = "get_group_list"         // 获取群列表
	ActionGetGroupMemberInfo  = "get_group_member_info"  // 获取群成员信息
	ActionGetGroupMemberList  = "get_group_member_list"  // 获取群成员列表
	ActionGetGroupHonorInfo   = "get_group_honor_info"   // 获取群荣誉信息
	ActionGetCookies          = "get_cookies"            // 获取Cookies
	ActionGetCsrfToken        = "get_csrf_token"         // 获取CSRF Token
	ActionGetCredentials      = "get_credentials"        // 获取QQ相关接口凭证
	ActionGetRecord           = "get_record"             // 获取语音
	ActionGetImage            = "get_image"              // 获取图片
	ActionCanSendImage        = "can_send_image"         // 检查是否可以发送图片
	ActionCanSendRecord       = "can_send_record"        // 检查是否可以发送语音
	ActionGetStatus           = "get_status"             // 获取运行状态
	ActionGetVersionInfo      = "get_version_info"       // 获取版本信息

	// 其他
	ActionSetRestart   = "set_restart"   // 重启OneBot实现
	ActionCleanCache   = "clean_cache"   // 清理缓存
	ActionSetQQProfile = ".set_qq_profile" // 设置登录号资料（需要扩展API）
)
