package adb

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

// findAvailablePort 查找一个可用的本地 TCP 端口
// 通过监听一个随机端口获取系统分配的可用端口
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("查找可用端口失败: %w", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// Forward 建立 ADB 端口转发
// serial: 设备序列号
// remote: 远程地址（如 "localabstract:scrcpy" 或 "tcp:5555"）
// 返回分配的本地端口号
func (c *Client) Forward(ctx context.Context, serial string, remote string) (int, error) {
	logInfo("建立端口转发 [%s]: %s", serial, remote)

	// 先检查是否已有该设备的转发
	existingPort, err := c.getExistingForward(ctx, serial, remote)
	if err == nil && existingPort > 0 {
		logDebug("复用已有转发，端口: %d", existingPort)
		return existingPort, nil
	}

	// 查找可用端口
	port, err := findAvailablePort()
	if err != nil {
		return 0, fmt.Errorf("查找可用端口失败: %w", err)
	}

	// 建立转发
	local := fmt.Sprintf("tcp:%d", port)
	_, err = c.runDeviceCommand(ctx, serial, "forward", local, remote)
	if err != nil {
		return 0, fmt.Errorf("建立端口转发失败: %w", err)
	}

	logInfo("端口转发成功: %s -> %s (本地端口: %d)", local, remote, port)
	return port, nil
}

// getExistingForward 检查是否已存在到指定远程地址的转发
// 如果存在，返回本地端口；否则返回 0
func (c *Client) getExistingForward(ctx context.Context, serial string, remote string) (int, error) {
	logDebug("检查已有转发 [%s]: %s", serial, remote)

	output, err := c.runDeviceCommand(ctx, serial, "forward", "--list")
	if err != nil {
		return 0, err
	}

	// 解析转发列表
	// 格式: "serial tcp:PORT remote"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[0] == serial && parts[2] == remote {
			// 提取端口号
			portStr := strings.TrimPrefix(parts[1], "tcp:")
			port, err := strconv.Atoi(portStr)
			if err == nil {
				return port, nil
			}
		}
	}

	return 0, nil
}

// RemoveForward 移除 ADB 端口转发
func (c *Client) RemoveForward(ctx context.Context, serial string, port int) error {
	logInfo("移除端口转发 [%s]: tcp:%d", serial, port)

	local := fmt.Sprintf("tcp:%d", port)
	_, err := c.runDeviceCommand(ctx, serial, "forward", "--remove", local)
	if err != nil {
		return fmt.Errorf("移除端口转发失败: %w", err)
	}

	logInfo("端口转发已移除: tcp:%d", port)
	return nil
}

// RemoveAllForwards 移除指定设备的所有端口转发
func (c *Client) RemoveAllForwards(ctx context.Context, serial string) error {
	logInfo("移除 [%s] 的所有端口转发", serial)

	_, err := c.runDeviceCommand(ctx, serial, "forward", "--remove-all")
	if err != nil {
		return fmt.Errorf("移除所有端口转发失败: %w", err)
	}

	logInfo("已移除 [%s] 的所有端口转发", serial)
	return nil
}

// Reverse 建立 ADB 反向隧道
// serial: 设备序列号
// remoteAbstract: 设备上的 abstract socket 名称 (如 "scrcpy")
// localPort: PC 本地监听端口
func (c *Client) Reverse(ctx context.Context, serial string, remoteAbstract string, localPort int) error {
	logInfo("建立反向隧道 [%s]: localabstract:%s -> tcp:%d", serial, remoteAbstract, localPort)

	remote := fmt.Sprintf("localabstract:%s", remoteAbstract)
	local := fmt.Sprintf("tcp:%d", localPort)

	_, err := c.runDeviceCommand(ctx, serial, "reverse", remote, local)
	if err != nil {
		return fmt.Errorf("建立反向隧道失败: %w", err)
	}

	logInfo("反向隧道建立成功: %s -> %s", remote, local)
	return nil
}

// RemoveAllReverses 移除指定设备的所有反向隧道
func (c *Client) RemoveAllReverses(ctx context.Context, serial string) error {
	logInfo("移除 [%s] 的所有反向隧道", serial)

	_, err := c.runDeviceCommand(ctx, serial, "reverse", "--remove-all")
	if err != nil {
		// 可能没有反向隧道，忽略错误
		logDebug("移除反向隧道结果: %v", err)
	}

	logInfo("已移除 [%s] 的所有反向隧道", serial)
	return nil
}

// Push 推送文件到设备
// local: 本地文件路径
// remote: 设备上的目标路径
func (c *Client) Push(ctx context.Context, serial string, local string, remote string) error {
	logInfo("推送文件 [%s]: %s -> %s", serial, local, remote)

	// 转换 Windows 路径分隔符为正斜杠
	local = filepath.ToSlash(local)

	_, err := c.runDeviceCommand(ctx, serial, "push", local, remote)
	if err != nil {
		return fmt.Errorf("推送文件失败: %w", err)
	}

	logInfo("文件推送成功: %s", remote)
	return nil
}

// Pull 从设备拉取文件
// remote: 设备上的文件路径
// local: 本地目标路径
func (c *Client) Pull(ctx context.Context, serial string, remote string, local string) error {
	logInfo("拉取文件 [%s]: %s -> %s", serial, remote, local)

	_, err := c.runDeviceCommand(ctx, serial, "pull", remote, local)
	if err != nil {
		return fmt.Errorf("拉取文件失败: %w", err)
	}

	logInfo("文件拉取成功: %s", local)
	return nil
}

// IsServerRunning 检查 scrcpy-server 是否在设备上运行
func (c *Client) IsServerRunning(ctx context.Context, serial string) (bool, error) {
	logDebug("检查 scrcpy-server 运行状态 [%s]", serial)

	output, err := c.Shell(ctx, serial, "ps -A | grep scrcpy")
	if err != nil {
		return false, err
	}

	return strings.Contains(output, "scrcpy"), nil
}

// KillServer 终止设备上的 scrcpy-server 进程
func (c *Client) KillServer(ctx context.Context, serial string) error {
	logInfo("终止 scrcpy-server [%s]", serial)

	_, err := c.Shell(ctx, serial, "pkill -f scrcpy")
	if err != nil {
		// 进程不存在时 pkill 可能返回错误，忽略
		logDebug("pkill 结果: %v (可能进程不存在)", err)
	}

	return nil
}
