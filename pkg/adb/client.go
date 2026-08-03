// Package adb 封装 ADB 命令，提供设备管理和连接功能
package adb

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/yuan71058/go-scrcpy/pkg/types"
)

// 日志级别常量
const (
	LogLevelNone = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

// 全局日志级别，可通过 SetLogLevel 调整
var logLevel = LogLevelError

// SetLogLevel 设置 ADB 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

// logDebug 输出调试级别日志
func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[ADB DEBUG] "+format+"\n", args...)
	}
}

// logInfo 输出信息级别日志
func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[ADB INFO] "+format+"\n", args...)
	}
}

// logError 输出错误级别日志
func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[ADB ERROR] "+format+"\n", args...)
	}
}

// Client 封装 ADB 命令行工具
type Client struct {
	ExecPath string // ADB 可执行文件路径，默认 "adb"
	mu       sync.Mutex
}

// NewClient 创建一个新的 ADB 客户端
// execPath 为空时默认使用系统 PATH 中的 "adb"
func NewClient(execPath string) *Client {
	if execPath == "" {
		execPath = "adb"
	}
	return &Client{
		ExecPath: execPath,
	}
}

// runCommand 执行 ADB 命令并返回输出
// serial 为空时执行全局命令（如 adb devices）
// serial 非空时执行指定设备命令（如 -s serial shell ...）
func (c *Client) runCommand(ctx context.Context, args ...string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logDebug("执行命令: %s %s", c.ExecPath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, c.ExecPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logError("命令执行失败: %v, stderr: %s", err, stderr.String())
		return "", fmt.Errorf("adb 命令失败: %w, stderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	logDebug("命令输出: %s", output)
	return output, nil
}

// runDeviceCommand 执行针对指定设备的 ADB 命令
func (c *Client) runDeviceCommand(ctx context.Context, serial string, args ...string) (string, error) {
	deviceArgs := []string{"-s", serial}
	deviceArgs = append(deviceArgs, args...)
	return c.runCommand(ctx, deviceArgs...)
}

// ListDevices 列举所有已连接的 ADB 设备
// 返回设备列表，包括序列号和状态
func (c *Client) ListDevices(ctx context.Context) ([]types.Device, error) {
	logInfo("列举已连接设备...")

	output, err := c.runCommand(ctx, "devices", "-l")
	if err != nil {
		return nil, fmt.Errorf("列举设备失败: %w", err)
	}

	var devices []types.Device
	lines := strings.Split(output, "\n")

	// 跳过第一行 "List of devices attached"
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		device := parseDeviceLine(line)
		if device.Serial != "" {
			devices = append(devices, device)
			logInfo("发现设备: %s (状态: %s)", device.Serial, device.State)
		}
	}

	logInfo("共发现 %d 个设备", len(devices))
	return devices, nil
}

// parseDeviceLine 解析 adb devices -l 输出的单行设备信息
// 格式: "SERIAL state product:model:... transport:..."
func parseDeviceLine(line string) types.Device {
	device := types.Device{}

	// 按空格分割
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return device
	}

	device.Serial = parts[0]
	device.State = parts[1]

	// 解析键值对（如 product:sunfire:model:Nexus_5）
	for i := 2; i < len(parts); i++ {
		kv := strings.SplitN(parts[i], ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "product":
			device.Product = kv[1]
		case "model":
			device.Model = kv[1]
		case "device":
			// device 字段可能重复，忽略
		case "transport_id":
			device.Transport = kv[1]
		}
	}

	return device
}

// IsDeviceConnected 检查指定设备是否已连接
func (c *Client) IsDeviceConnected(ctx context.Context, serial string) (bool, error) {
	logDebug("检查设备连接状态: %s", serial)

	_, err := c.runDeviceCommand(ctx, serial, "get-state")
	if err != nil {
		// 设备不存在或离线时会返回错误
		return false, nil
	}
	return true, nil
}

// GetDeviceModel 获取设备型号
func (c *Client) GetDeviceModel(ctx context.Context, serial string) (string, error) {
	logDebug("获取设备型号: %s", serial)

	output, err := c.runDeviceCommand(ctx, serial, "shell", "getprop", "ro.product.model")
	if err != nil {
		return "", fmt.Errorf("获取设备型号失败: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// GetAndroidVersion 获取 Android 版本号
func (c *Client) GetAndroidVersion(ctx context.Context, serial string) (int, error) {
	logDebug("获取 Android 版本: %s", serial)

	output, err := c.runDeviceCommand(ctx, serial, "shell", "getprop", "ro.build.version.sdk")
	if err != nil {
		return 0, fmt.Errorf("获取 Android 版本失败: %w", err)
	}

	version, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("解析版本号失败: %w", err)
	}
	return version, nil
}

// GetDeviceProperties 获取设备所有系统属性
func (c *Client) GetDeviceProperties(ctx context.Context, serial string) (map[string]string, error) {
	logDebug("获取设备属性: %s", serial)

	output, err := c.runDeviceCommand(ctx, serial, "shell", "getprop")
	if err != nil {
		return nil, fmt.Errorf("获取设备属性失败: %w", err)
	}

	props := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 格式: "[key]: [value]"
		if strings.HasPrefix(line, "[") {
			parts := strings.SplitN(line, "]: [", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], "[")
				value := strings.TrimSuffix(parts[1], "]")
				props[key] = value
			}
		}
	}

	return props, nil
}

// Shell 在指定设备上执行 shell 命令
func (c *Client) Shell(ctx context.Context, serial string, command string) (string, error) {
	logDebug("执行 shell 命令 [%s]: %s", serial, command)

	output, err := c.runDeviceCommand(ctx, serial, "shell", command)
	if err != nil {
		return "", fmt.Errorf("执行 shell 命令失败: %w", err)
	}
	return output, nil
}
