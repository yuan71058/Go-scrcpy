// Package transport 管理与 scrcpy-server 的三通道 socket 连接
package transport

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// 日志级别常量
const (
	LogLevelNone = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

var logLevel = LogLevelError

// SetLogLevel 设置 transport 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[TRANSPORT DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[TRANSPORT INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[TRANSPORT ERROR] "+format+"\n", args...)
	}
}

// Connection 封装与 scrcpy-server 的三通道连接
type Connection struct {
	VideoConn   net.Conn // 视频通道连接
	AudioConn   net.Conn // 音频通道连接
	ControlConn net.Conn // 控制通道连接
	mu          sync.Mutex
}

// NewConnection 创建三通道连接
// videoPort, audioPort, controlPort: 本地转发端口
func NewConnection(videoPort, audioPort, controlPort int) (*Connection, error) {
	logInfo("建立三通道连接: video=%d, audio=%d, control=%d", videoPort, audioPort, controlPort)

	// 带重试的连接函数
	dialWithRetry := func(port int) (net.Conn, error) {
		var lastErr error
		for i := 0; i < 10; i++ {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
			if err == nil {
				return conn, nil
			}
			lastErr = err
			logDebug("连接端口 %d 失败 (第 %d 次): %v", port, i+1, err)
			time.Sleep(500 * time.Millisecond)
		}
		return nil, lastErr
	}

	// 连接视频通道
	videoConn, err := dialWithRetry(videoPort)
	if err != nil {
		return nil, fmt.Errorf("连接视频通道失败: %w", err)
	}
	logDebug("视频通道已连接")

	// 连接音频通道
	audioConn, err := dialWithRetry(audioPort)
	if err != nil {
		videoConn.Close()
		return nil, fmt.Errorf("连接音频通道失败: %w", err)
	}
	logDebug("音频通道已连接")

	// 连接控制通道
	controlConn, err := dialWithRetry(controlPort)
	if err != nil {
		videoConn.Close()
		audioConn.Close()
		return nil, fmt.Errorf("连接控制通道失败: %w", err)
	}
	logDebug("控制通道已连接")

	logInfo("三通道连接建立成功")
	return &Connection{
		VideoConn:   videoConn,
		AudioConn:   audioConn,
		ControlConn: controlConn,
	}, nil
}

// Close 关闭所有通道连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error

	if c.VideoConn != nil {
		if err := c.VideoConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		c.VideoConn = nil
	}

	if c.AudioConn != nil {
		if err := c.AudioConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		c.AudioConn = nil
	}

	if c.ControlConn != nil {
		if err := c.ControlConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		c.ControlConn = nil
	}

	logDebug("三通道连接已关闭")
	return firstErr
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.VideoConn == nil && c.AudioConn == nil && c.ControlConn == nil
}
