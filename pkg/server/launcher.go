package server

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
)

// 日志级别常量
const (
	LogLevelNone = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

// 全局日志级别
var logLevel = LogLevelError

// SetLogLevel 设置 server 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[SERVER DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[SERVER INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[SERVER ERROR] "+format+"\n", args...)
	}
}

// ServerConn 表示与 scrcpy-server 的连接信息
type ServerConn struct {
	Serial    string // 设备序列号
	VideoPort int    // 视频通道本地端口
	AudioPort int    // 音频通道本地端口
	ControlPort int  // 控制通道本地端口
}

// Launcher 管理 scrcpy-server 的生命周期
type Launcher struct {
	ADB *adb.Client // ADB 客户端
}

// NewLauncher 创建 server 启动器
func NewLauncher(adbClient *adb.Client) *Launcher {
	return &Launcher{
		ADB: adbClient,
	}
}

// Start 启动 scrcpy-server 并建立连接
// serial: 设备序列号
// opts: server 启动参数
// localJAR: 本地 scrcpy-server.jar 路径（为空则使用设备上已有的）
func (l *Launcher) Start(ctx context.Context, serial string, opts Options, localJAR string) (*ServerConn, error) {
	logInfo("启动 scrcpy-server [%s]...", serial)

	// 检查 server 是否已在运行
	running, err := l.ADB.IsServerRunning(ctx, serial)
	if err != nil {
		logDebug("检查 server 状态失败: %v", err)
	}
	if running {
		logInfo("scrcpy-server 已在运行 [%s]，先终止旧进程", serial)
		if err := l.ADB.KillServer(ctx, serial); err != nil {
			logDebug("终止旧 server 失败: %v", err)
		}
	}

	// 推送 JAR 到设备
	if localJAR != "" {
		if err := l.pushServer(ctx, serial, localJAR); err != nil {
			return nil, fmt.Errorf("推送 server 失败: %w", err)
		}
	}

	// 构建启动参数
	args := opts.ToArgs()
	logInfo("server 参数: %s", strings.Join(args, " "))

	// 启动 server
	if err := l.startServer(ctx, serial, args); err != nil {
		return nil, fmt.Errorf("启动 server 失败: %w", err)
	}

	// 建立端口转发
	videoPort, err := l.ADB.Forward(ctx, serial, "localabstract:scrcpy")
	if err != nil {
		return nil, fmt.Errorf("视频端口转发失败: %w", err)
	}

	audioPort, err := l.ADB.Forward(ctx, serial, "localabstract:scrcpy_1")
	if err != nil {
		return nil, fmt.Errorf("音频端口转发失败: %w", err)
	}

	controlPort, err := l.ADB.Forward(ctx, serial, "localabstract:scrcpy_2")
	if err != nil {
		return nil, fmt.Errorf("控制端口转发失败: %w", err)
	}

	conn := &ServerConn{
		Serial:      serial,
		VideoPort:   videoPort,
		AudioPort:   audioPort,
		ControlPort: controlPort,
	}

	logInfo("scrcpy-server 启动成功 [%s]: video=%d, audio=%d, control=%d",
		serial, videoPort, audioPort, controlPort)

	return conn, nil
}

// pushServer 推送 scrcpy-server.jar 到设备
func (l *Launcher) pushServer(ctx context.Context, serial string, localJAR string) error {
	logInfo("推送 scrcpy-server.jar [%s]: %s -> %s", serial, localJAR, ServerPath)

	// 确保本地文件存在
	absPath, err := filepath.Abs(localJAR)
	if err != nil {
		return fmt.Errorf("解析本地 JAR 路径失败: %w", err)
	}

	if err := l.ADB.Push(ctx, serial, absPath, ServerPath); err != nil {
		return fmt.Errorf("推送 JAR 文件失败: %w", err)
	}

	return nil
}

// startServer 通过 adb shell 启动 scrcpy-server
func (l *Launcher) startServer(ctx context.Context, serial string, args []string) error {
	logDebug("执行 server 启动命令 [%s]", serial)

	// 构建启动命令
	// 格式: CLASSPATH=/data/local/tmp/scrcpy-server.jar nohup app_process / com.genymobile.scrcpy.Server <args>
	classpath := fmt.Sprintf("CLASSPATH=%s", ServerPath)
	serverArgs := strings.Join(args, " ")
	command := fmt.Sprintf("%s nohup app_process / com.genymobile.scrcpy.Server %s >/dev/null 2>&1 &",
		classpath, serverArgs)

	logDebug("启动命令: %s", command)

	// 执行启动命令
	_, err := l.ADB.Shell(ctx, serial, command)
	if err != nil {
		return fmt.Errorf("执行启动命令失败: %w", err)
	}

	// 等待 server 启动并监听 socket
	logDebug("等待 server 启动...")
	if err := l.waitForServer(ctx, serial, 5*time.Second); err != nil {
		return fmt.Errorf("等待 server 启动超时: %w", err)
	}

	return nil
}

// waitForServer 等待 scrcpy-server 启动并创建 socket
func (l *Launcher) waitForServer(ctx context.Context, serial string, timeout time.Duration) error {
	logDebug("等待 server 启动...")
	// 等待 server 进程初始化和创建 socket
	// 使用固定延迟，因为 ps 命令在部分设备上不可用
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}
	return nil
}

// Kill 终止指定设备上的 scrcpy-server
func (l *Launcher) Kill(ctx context.Context, serial string) error {
	logInfo("终止 scrcpy-server [%s]", serial)

	// 移除端口转发
	if err := l.ADB.RemoveAllForwards(ctx, serial); err != nil {
		logDebug("移除端口转发失败: %v", err)
	}

	// 终止 server 进程
	if err := l.ADB.KillServer(ctx, serial); err != nil {
		logDebug("终止 server 进程失败: %v", err)
	}

	logInfo("scrcpy-server 已终止 [%s]", serial)
	return nil
}

// IsRunning 检查 scrcpy-server 是否在运行
func (l *Launcher) IsRunning(ctx context.Context, serial string) (bool, error) {
	return l.ADB.IsServerRunning(ctx, serial)
}
