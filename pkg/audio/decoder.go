// Package audio 提供音频解码和播放功能
package audio

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

// SetLogLevel 设置 audio 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[AUDIO DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[AUDIO INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[AUDIO ERROR] "+format+"\n", args...)
	}
}

// Decoder 音频解码器接口
type Decoder interface {
	// Push 推送编码数据
	Push(data []byte) error
	// ReadPCM 读取解码后的 PCM 数据（阻塞）
	ReadPCM() (*AudioData, error)
	// Close 关闭解码器
	Close() error
}

// Player 音频播放器接口
type Player interface {
	// Play 播放 PCM 数据
	Play(pcm []byte) error
	// Close 关闭播放器
	Close() error
}

// AudioData 解码后的音频数据
type AudioData struct {
	PTS        int64  // 显示时间戳 (微秒)
	Data       []byte // PCM 数据
	SampleRate int    // 采样率
	Channels   int    // 声道数
	BitsPerSample int // 位深度
}
