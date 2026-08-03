// Package transport 管理与 scrcpy-server 的三通道 socket 连接
// 使用 reverse tunnel 模式：Android 服务端主动连接 PC
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

// Listener 管理 reverse tunnel 监听器
// Android 服务端主动连接 PC
type Listener struct {
	ln        net.Listener
	videoConn net.Conn
	audioConn net.Conn
	controlConn net.Conn
	mu        sync.Mutex
}

// NewListener 创建 reverse tunnel 监听器
// port: PC 监听端口
func NewListener(port int) (*Listener, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("监听端口 %d 失败: %w", port, err)
	}

	logInfo("reverse tunnel 监听器已启动: %s", addr)
	return &Listener{ln: ln}, nil
}

// GetPort 获取监听端口
func (l *Listener) GetPort() int {
	return l.ln.Addr().(*net.TCPAddr).Port
}

// Accept 等待 Android 服务端连接
// count: 需要接受的连接数量 (1-3)
// timeout: 超时时间
func (l *Listener) Accept(count int, timeout time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	deadline := time.Now().Add(timeout)

	for i := 0; i < count; i++ {
		if time.Now().After(deadline) {
			return fmt.Errorf("等待连接超时")
		}

		// 设置接受超时
		ln := l.ln.(*net.TCPListener)
		ln.SetDeadline(deadline)

		conn, err := l.ln.Accept()
		if err != nil {
			return fmt.Errorf("接受第 %d 条连接失败: %w", i+1, err)
		}

		switch i {
		case 0:
			l.videoConn = conn
			logDebug("视频通道已连接: %s", conn.RemoteAddr())
		case 1:
			l.audioConn = conn
			logDebug("音频通道已连接: %s", conn.RemoteAddr())
		case 2:
			l.controlConn = conn
			logDebug("控制通道已连接: %s", conn.RemoteAddr())
		}
	}

	// 清除超时
	ln := l.ln.(*net.TCPListener)
	ln.SetDeadline(time.Time{})

	return nil
}

// GetVideoConn 获取视频通道连接
func (l *Listener) GetVideoConn() net.Conn {
	return l.videoConn
}

// GetAudioConn 获取音频通道连接
func (l *Listener) GetAudioConn() net.Conn {
	return l.audioConn
}

// GetControlConn 获取控制通道连接
func (l *Listener) GetControlConn() net.Conn {
	return l.controlConn
}

// Close 关闭监听器和所有连接
func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var firstErr error

	if l.videoConn != nil {
		if err := l.videoConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if l.audioConn != nil {
		if err := l.audioConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if l.controlConn != nil {
		if err := l.controlConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if l.ln != nil {
		if err := l.ln.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	logDebug("reverse tunnel 监听器已关闭")
	return firstErr
}

// Connection 封装与 scrcpy-server 的三通道连接 (兼容旧接口)
type Connection struct {
	VideoConn   net.Conn
	AudioConn   net.Conn
	ControlConn net.Conn
	mu          sync.Mutex
}

// NewConnection 从 Listener 创建 Connection
func NewConnectionFromListener(l *Listener) *Connection {
	return &Connection{
		VideoConn:   l.GetVideoConn(),
		AudioConn:   l.GetAudioConn(),
		ControlConn: l.GetControlConn(),
	}
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
