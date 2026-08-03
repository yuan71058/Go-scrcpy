// Package record 提供屏幕录制和截图功能
package record

import (
	"fmt"
	"io"
	"os"
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

// SetLogLevel 设置 record 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[RECORD DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[RECORD INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[RECORD ERROR] "+format+"\n", args...)
	}
}

// Recorder 屏幕录制器
type Recorder struct {
	output    io.Writer
	format    string
	mu        sync.Mutex
	closed    bool
	frameCount int64
}

// NewRecorder 创建录制器
// output: 输出流（文件、网络等）
// format: 格式 ("mp4", "mkv")
func NewRecorder(output io.Writer, format string) *Recorder {
	if format == "" {
		format = "mp4"
	}
	return &Recorder{
		output: output,
		format: format,
	}
}

// NewRecorderFromFile 从文件路径创建录制器
func NewRecorderFromFile(path string, format string) (*Recorder, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("创建录制文件失败: %w", err)
	}

	return NewRecorder(file, format), nil
}

// WriteVideo 写入视频帧
// data: 编码数据
// pts: 显示时间戳 (微秒)
// keyframe: 是否为关键帧
func (r *Recorder) WriteVideo(data []byte, pts int64, keyframe bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("录制器已关闭")
	}

	// 简单实现：将数据写入输出流
	// 实际应用中应使用 MP4/MKV 容器封装
	_, err := r.output.Write(data)
	if err != nil {
		return fmt.Errorf("写入视频数据失败: %w", err)
	}

	r.frameCount++
	logDebug("写入视频帧: pts=%d, key=%v, size=%d, total=%d", pts, keyframe, len(data), r.frameCount)
	return nil
}

// WriteAudio 写入音频数据
// data: 编码数据
// pts: 显示时间戳 (微秒)
func (r *Recorder) WriteAudio(data []byte, pts int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("录制器已关闭")
	}

	// 简单实现：将数据写入输出流
	_, err := r.output.Write(data)
	if err != nil {
		return fmt.Errorf("写入音频数据失败: %w", err)
	}

	logDebug("写入音频数据: pts=%d, size=%d", pts, len(data))
	return nil
}

// GetFrameCount 获取已录制帧数
func (r *Recorder) GetFrameCount() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.frameCount
}

// Close 关闭录制器
func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	logInfo("录制器已关闭，共录制 %d 帧", r.frameCount)
	return nil
}
