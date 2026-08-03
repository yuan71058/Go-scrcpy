// Package video 提供视频解码和渲染功能
package video

import (
	"fmt"
)

// 日志级别常量
const (
	LogLevelNone = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

var logLevel = LogLevelError

// SetLogLevel 设置 video 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[VIDEO DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[VIDEO INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[VIDEO ERROR] "+format+"\n", args...)
	}
}

// Decoder 视频解码器接口
// 用户可替换为自定义实现（如 FFmpeg CGo 绑定）
type Decoder interface {
	// Push 推送编码数据到解码器
	Push(data []byte) error
	// ReadFrame 读取解码后的帧（阻塞）
	ReadFrame() (*DecodedFrame, error)
	// Close 关闭解码器
	Close() error
}

// DecodedFrame 解码后的视频帧
type DecodedFrame struct {
	PTS    int64  // 显示时间戳 (微秒)
	Data   []byte // 原始编码数据（NALU）
	Width  int    // 帧宽度
	Height int    // 帧高度
	Config bool   // 是否为配置帧 (SPS/PPS)
	Key    bool   // 是否为关键帧
}

// Renderer 渲染器接口
type Renderer interface {
	// Render 渲染一帧
	Render(frame *DecodedFrame) error
	// Close 关闭渲染器
	Close() error
}
