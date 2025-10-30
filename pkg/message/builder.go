package message

import (
	"fmt"
	types "onebot-go2/pkg/const"
	"strconv"
)

// Builder 消息构造器，支持链式调用
type Builder struct {
	messages types.MessageArray
}

// NewBuilder 创建新的消息构造器
func NewBuilder() *Builder {
	return &Builder{
		messages: make(types.MessageArray, 0),
	}
}

// Text 添加文本消息段
func (b *Builder) Text(text string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "text",
		Data: map[string]interface{}{
			"text": text,
		},
	})
	return b
}

// Face 添加QQ表情消息段
func (b *Builder) Face(id int) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "face",
		Data: map[string]interface{}{
			"id": strconv.Itoa(id),
		},
	})
	return b
}

// Image 添加图片消息段
// file 可以是文件路径、URL、base64等
func (b *Builder) Image(file string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "image",
		Data: map[string]interface{}{
			"file": file,
		},
	})
	return b
}

// ImageWithCache 添加图片消息段（可控制缓存）
func (b *Builder) ImageWithCache(file string, useCache bool) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "image",
		Data: map[string]interface{}{
			"file":  file,
			"cache": useCache,
		},
	})
	return b
}

// Record 添加语音消息段
func (b *Builder) Record(file string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "record",
		Data: map[string]interface{}{
			"file": file,
		},
	})
	return b
}

// Video 添加视频消息段
func (b *Builder) Video(file string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "video",
		Data: map[string]interface{}{
			"file": file,
		},
	})
	return b
}

// At 添加@某人消息段
func (b *Builder) At(userID int64) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "at",
		Data: map[string]interface{}{
			"qq": strconv.FormatInt(userID, 10),
		},
	})
	return b
}

// AtAll 添加@全体成员消息段
func (b *Builder) AtAll() *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "at",
		Data: map[string]interface{}{
			"qq": "all",
		},
	})
	return b
}

// Reply 添加回复消息段
func (b *Builder) Reply(messageID int32) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "reply",
		Data: map[string]interface{}{
			"id": strconv.Itoa(int(messageID)),
		},
	})
	return b
}

// Poke 添加戳一戳消息段
func (b *Builder) Poke(userID int64) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "poke",
		Data: map[string]interface{}{
			"qq": strconv.FormatInt(userID, 10),
		},
	})
	return b
}

// Share 添加链接分享消息段
func (b *Builder) Share(url, title string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "share",
		Data: map[string]interface{}{
			"url":   url,
			"title": title,
		},
	})
	return b
}

// ShareWithDetail 添加链接分享消息段（带详细信息）
func (b *Builder) ShareWithDetail(url, title, content, image string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "share",
		Data: map[string]interface{}{
			"url":     url,
			"title":   title,
			"content": content,
			"image":   image,
		},
	})
	return b
}

// Music 添加音乐分享消息段
// musicType: qq, 163, xm (QQ音乐、网易云音乐、虾米音乐)
func (b *Builder) Music(musicType string, id int64) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "music",
		Data: map[string]interface{}{
			"type": musicType,
			"id":   strconv.FormatInt(id, 10),
		},
	})
	return b
}

// CustomMusic 添加自定义音乐分享消息段
func (b *Builder) CustomMusic(url, audio, title string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "music",
		Data: map[string]interface{}{
			"type":  "custom",
			"url":   url,
			"audio": audio,
			"title": title,
		},
	})
	return b
}

// Forward 添加合并转发消息段
func (b *Builder) Forward(id string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "forward",
		Data: map[string]interface{}{
			"id": id,
		},
	})
	return b
}

// Node 添加合并转发节点消息段
func (b *Builder) Node(userID int64, nickname string, content types.MessageArray) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "node",
		Data: map[string]interface{}{
			"user_id":  strconv.FormatInt(userID, 10),
			"nickname": nickname,
			"content":  content,
		},
	})
	return b
}

// Json 添加JSON消息段
func (b *Builder) Json(data string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "json",
		Data: map[string]interface{}{
			"data": data,
		},
	})
	return b
}

// Xml 添加XML消息段
func (b *Builder) Xml(data string) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: "xml",
		Data: map[string]interface{}{
			"data": data,
		},
	})
	return b
}

// Custom 添加自定义消息段
func (b *Builder) Custom(msgType string, data map[string]interface{}) *Builder {
	b.messages = append(b.messages, types.Message{
		Type: msgType,
		Data: data,
	})
	return b
}

// Build 构建消息数组
func (b *Builder) Build() types.MessageArray {
	return b.messages
}

// Clear 清空消息构造器
func (b *Builder) Clear() *Builder {
	b.messages = make(types.MessageArray, 0)
	return b
}

// 快捷构造函数

// Text 快捷创建纯文本消息
func Text(text string) types.MessageArray {
	return NewBuilder().Text(text).Build()
}

// Image 快捷创建图片消息
func Image(file string) types.MessageArray {
	return NewBuilder().Image(file).Build()
}

// At 快捷创建@消息
func At(userID int64) types.MessageArray {
	return NewBuilder().At(userID).Build()
}

// Reply 快捷创建回复消息
func Reply(messageID int32, text string) types.MessageArray {
	return NewBuilder().Reply(messageID).Text(text).Build()
}

// AtText 快捷创建@+文本消息
func AtText(userID int64, text string) types.MessageArray {
	return NewBuilder().At(userID).Text(" ").Text(text).Build()
}

// ImageText 快捷创建图片+文本消息
func ImageText(imageFile, text string) types.MessageArray {
	return NewBuilder().Image(imageFile).Text(text).Build()
}

// String 将消息数组转换为纯文本（用于日志等）
func String(messages types.MessageArray) string {
	result := ""
	for _, msg := range messages {
		switch msg.Type {
		case "text":
			if text, ok := msg.Data["text"].(string); ok {
				result += text
			}
		case "face":
			if id, ok := msg.Data["id"].(string); ok {
				result += fmt.Sprintf("[表情:%s]", id)
			}
		case "image":
			result += "[图片]"
		case "record":
			result += "[语音]"
		case "video":
			result += "[视频]"
		case "at":
			if qq, ok := msg.Data["qq"].(string); ok {
				if qq == "all" {
					result += "@全体成员"
				} else {
					result += fmt.Sprintf("@%s", qq)
				}
			}
		case "reply":
			if id, ok := msg.Data["id"].(string); ok {
				result += fmt.Sprintf("[回复:%s]", id)
			}
		default:
			result += fmt.Sprintf("[%s]", msg.Type)
		}
	}
	return result
}
