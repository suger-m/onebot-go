package handler

import (
	"fmt"
	"log"
	"strings"

	types "onebot-go2/pkg/const"
	"onebot-go2/pkg/event"
	"onebot-go2/pkg/message"
)

// CommandHandler 命令处理器
// 支持命令路由和参数解析
type CommandHandler struct {
	prefix   string                                                 // 命令前缀，如 "/"
	commands map[string]func(*event.Context[*types.MessageEvent]) error // 命令处理函数映射
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler(prefix string) *CommandHandler {
	return &CommandHandler{
		prefix:   prefix,
		commands: make(map[string]func(*event.Context[*types.MessageEvent]) error),
	}
}

// Register 注册命令
func (h *CommandHandler) Register(command string, handler func(*event.Context[*types.MessageEvent]) error) {
	h.commands[command] = handler
	log.Printf("[CommandHandler] Registered command: %s%s", h.prefix, command)
}

// Handle 处理消息事件
func (h *CommandHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
	rawMsg := ctx.Event.RawMessage

	// 检查是否以命令前缀开头
	if !strings.HasPrefix(rawMsg, h.prefix) {
		return nil
	}

	// 去除前缀
	cmdLine := strings.TrimPrefix(rawMsg, h.prefix)

	// 解析命令和参数
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	// 存储命令和参数到上下文，供后续使用
	ctx.Set("command", command)
	ctx.Set("args", args)

	// 查找并执行命令处理函数
	if handler, exists := h.commands[command]; exists {
		log.Printf("[CommandHandler] Executing command: %s%s (args: %v)", h.prefix, command, args)
		return handler(ctx)
	}

	return nil
}

func (h *CommandHandler) Priority() int {
	return 30 // 较高优先级，在过滤器之后
}

func (h *CommandHandler) Name() string {
	return "CommandHandler"
}

// ============ 示例命令实现 ============

// HelpCommand /help 命令
func HelpCommand(ctx *event.Context[*types.MessageEvent]) error {
	helpText := `可用命令：
/help - 显示帮助信息
/ping - 测试响应
/echo <文本> - 回复相同文本
/info - 显示群/用户信息
/ban <@用户> <时长(秒)> - 禁言用户
/unban <@用户> - 解除禁言`

	_, err := ctx.ReplyText(helpText)
	return err
}

// PingCommand /ping 命令
func PingCommand(ctx *event.Context[*types.MessageEvent]) error {
	_, err := ctx.ReplyText("Pong!")
	return err
}

// EchoCommand /echo 命令
func EchoCommand(ctx *event.Context[*types.MessageEvent]) error {
	args, _ := ctx.Get("args")
	argList := args.([]string)

	if len(argList) == 0 {
		_, err := ctx.ReplyText("请提供要回复的内容")
		return err
	}

	text := strings.Join(argList, " ")
	_, err := ctx.ReplyText(text)
	return err
}

// InfoCommand /info 命令 - 显示群或用户信息
func InfoCommand(ctx *event.Context[*types.MessageEvent]) error {
	if ctx.IsGroupMessage() {
		groupID, _ := ctx.GetGroupID()
		userID, _ := ctx.GetUserID()

		// 获取群信息
		groupInfo, err := ctx.GetGroupInfo(groupID)
		if err != nil {
			_, _ = ctx.ReplyText(fmt.Sprintf("获取群信息失败: %v", err))
			return err
		}

		// 获取发送者信息
		memberInfo, err := ctx.GetGroupMemberInfo(groupID, userID)
		if err != nil {
			_, _ = ctx.ReplyText(fmt.Sprintf("获取成员信息失败: %v", err))
			return err
		}

		info := fmt.Sprintf(
			"群信息：\n群号：%d\n群名：%s\n成员数：%d/%d\n\n发送者：\n昵称：%s\n群名片：%s\n角色：%s",
			groupInfo.GroupID,
			groupInfo.GroupName,
			groupInfo.MemberCount,
			groupInfo.MaxMemberCount,
			memberInfo.Nickname,
			memberInfo.Card,
			memberInfo.Role,
		)

		_, err = ctx.ReplyText(info)
		return err
	} else if ctx.IsPrivateMessage() {
		userID, _ := ctx.GetUserID()
		info := fmt.Sprintf("私聊消息\n发送者：%d\n昵称：%s",
			userID,
			ctx.Event.Sender.Nickname,
		)
		_, err := ctx.ReplyText(info)
		return err
	}

	return nil
}

// BanCommand /ban 命令 - 禁言用户（仅管理员）
func BanCommand(ctx *event.Context[*types.MessageEvent]) error {
	if !ctx.IsGroupMessage() {
		_, err := ctx.ReplyText("该命令仅在群聊中可用")
		return err
	}

	groupID, _ := ctx.GetGroupID()
	args, _ := ctx.Get("args")
	argList := args.([]string)

	if len(argList) < 2 {
		_, err := ctx.ReplyText("用法: /ban <@用户> <时长(秒)>")
		return err
	}

	// 解析@的用户ID（简化版本，实际应该从消息段中解析）
	// 这里假设第一个参数是用户ID
	var targetUserID int64
	fmt.Sscanf(argList[0], "%d", &targetUserID)

	if targetUserID == 0 {
		_, err := ctx.ReplyText("请@要禁言的用户")
		return err
	}

	// 解析禁言时长
	var duration int64
	fmt.Sscanf(argList[1], "%d", &duration)

	if duration <= 0 {
		_, err := ctx.ReplyText("禁言时长必须大于0")
		return err
	}

	// 执行禁言
	if err := ctx.BanGroupMember(groupID, targetUserID, duration); err != nil {
		_, _ = ctx.ReplyText(fmt.Sprintf("禁言失败: %v", err))
		return err
	}

	_, err := ctx.ReplyText(fmt.Sprintf("已禁言用户 %d，时长 %d 秒", targetUserID, duration))
	return err
}

// UnbanCommand /unban 命令 - 解除禁言（仅管理员）
func UnbanCommand(ctx *event.Context[*types.MessageEvent]) error {
	if !ctx.IsGroupMessage() {
		_, err := ctx.ReplyText("该命令仅在群聊中可用")
		return err
	}

	groupID, _ := ctx.GetGroupID()
	args, _ := ctx.Get("args")
	argList := args.([]string)

	if len(argList) < 1 {
		_, err := ctx.ReplyText("用法: /unban <@用户>")
		return err
	}

	// 解析@的用户ID
	var targetUserID int64
	fmt.Sscanf(argList[0], "%d", &targetUserID)

	if targetUserID == 0 {
		_, err := ctx.ReplyText("请@要解除禁言的用户")
		return err
	}

	// 解除禁言
	if err := ctx.UnbanGroupMember(groupID, targetUserID); err != nil {
		_, _ = ctx.ReplyText(fmt.Sprintf("解除禁言失败: %v", err))
		return err
	}

	_, err := ctx.ReplyText(fmt.Sprintf("已解除用户 %d 的禁言", targetUserID))
	return err
}

// QuoteCommand /quote 命令 - 引用回复
func QuoteCommand(ctx *event.Context[*types.MessageEvent]) error {
	args, _ := ctx.Get("args")
	argList := args.([]string)

	if len(argList) == 0 {
		_, err := ctx.ReplyText("请提供要回复的内容")
		return err
	}

	text := strings.Join(argList, " ")
	msg := message.NewBuilder().Text(text).Build()

	_, err := ctx.ReplyWithQuote(msg)
	return err
}

// ImageCommand /image 命令 - 发送图片（示例）
func ImageCommand(ctx *event.Context[*types.MessageEvent]) error {
	args, _ := ctx.Get("args")
	argList := args.([]string)

	if len(argList) == 0 {
		_, err := ctx.ReplyText("用法: /image <图片URL>")
		return err
	}

	imageURL := argList[0]
	msg := message.NewBuilder().Image(imageURL).Text("这是你要的图片").Build()

	_, err := ctx.Reply(msg)
	return err
}

// ============ 创建默认命令处理器的辅助函数 ============

// NewDefaultCommandHandler 创建带有默认命令的命令处理器
func NewDefaultCommandHandler() *CommandHandler {
	handler := NewCommandHandler("/")

	// 注册默认命令
	handler.Register("help", HelpCommand)
	handler.Register("ping", PingCommand)
	handler.Register("echo", EchoCommand)
	handler.Register("info", InfoCommand)
	handler.Register("ban", BanCommand)
	handler.Register("unban", UnbanCommand)
	handler.Register("quote", QuoteCommand)
	handler.Register("image", ImageCommand)

	return handler
}
