// Package control 提供高级控制功能
// 包括剪贴板同步和文件推送
package control

import (
	"fmt"
	"sync"
)

// 日志级别常量
const (
	LogLevelNone = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

var logLevel = LogLevelError

// SetLogLevel 设置 control 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[CONTROL DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[CONTROL INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[CONTROL ERROR] "+format+"\n", args...)
	}
}

// Sender 控制消息发送器接口
type Sender interface {
	SendControl(msg []byte) error
}

// Clipboard 剪贴板管理器
type Clipboard struct {
	sender   Sender
	sequence uint64
	mu       sync.Mutex
	onChange func(text string)
	closed   bool
}

// NewClipboard 创建剪贴板管理器
func NewClipboard(sender Sender) *Clipboard {
	return &Clipboard{
		sender: sender,
	}
}

// OnChange 设置剪贴板变化回调
func (c *Clipboard) OnChange(fn func(text string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onChange = fn
}

// HandleMessage 处理设备消息
// 从控制通道收到的设备消息应转发到此方法
func (c *Clipboard) HandleMessage(msgType byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("剪贴板管理器已关闭")
	}

	switch msgType {
	case 0: // TYPE_CLIPBOARD
		text := string(payload)
		logInfo("剪贴板变化: %s", text)

		if c.onChange != nil {
			c.onChange(text)
		}
	case 1: // TYPE_ACK_CLIPBOARD
		if len(payload) >= 8 {
			sequence := uint64(payload[0])<<56 | uint64(payload[1])<<48 |
				uint64(payload[2])<<40 | uint64(payload[3])<<32 |
				uint64(payload[4])<<24 | uint64(payload[5])<<16 |
				uint64(payload[6])<<8 | uint64(payload[7])
			logDebug("剪贴板确认: sequence=%d", sequence)
		}
	default:
		logDebug("未知消息类型: %d", msgType)
	}

	return nil
}

// Close 关闭剪贴板管理器
func (c *Clipboard) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

// FilePusher 文件推送器
type FilePusher struct {
	sender     Sender
	pushID     int16
	mu         sync.Mutex
	closed     bool
}

// NewFilePusher 创建文件推送器
func NewFilePusher(sender Sender) *FilePusher {
	return &FilePusher{
		sender: sender,
	}
}

// PushProgress 推送进度
type PushProgress struct {
	ID        int16  // 推送 ID
	Code      byte   // 状态码
	BytesSent int64  // 已发送字节数
	Total     int64  // 总字节数
}

// HandleMessage 处理设备消息
func (f *FilePusher) HandleMessage(msgType byte, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return fmt.Errorf("文件推送器已关闭")
	}

	if msgType == 101 { // TYPE_FILE_PUSH_RESPONSE
		if len(payload) >= 3 {
			id := int16(payload[0])<<8 | int16(payload[1])
			code := payload[2]
			logInfo("文件推送响应: id=%d, code=%d", id, code)
		}
	}

	return nil
}

// Close 关闭文件推送器
func (f *FilePusher) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}
