package adb

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/yuan71058/go-scrcpy/pkg/types"
)

// DeviceTracker 设备跟踪器，用于实时监测设备上下线
type DeviceTracker struct {
	client     *Client
	onAdd      func(device types.Device)
	onRemove   func(device types.Device)
	onChange   func(device types.Device)
	mu         sync.RWMutex
	running    bool
	cancelFunc context.CancelFunc
}

// NewDeviceTracker 创建设备跟踪器
func NewDeviceTracker(client *Client) *DeviceTracker {
	return &DeviceTracker{
		client: client,
	}
}

// OnAdd 设置设备上线回调
func (t *DeviceTracker) OnAdd(fn func(device types.Device)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onAdd = fn
}

// OnRemove 设置设备离线回调
func (t *DeviceTracker) OnRemove(fn func(device types.Device)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onRemove = fn
}

// OnChange 设置设备状态变化回调
func (t *DeviceTracker) OnChange(fn func(device types.Device)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onChange = fn
}

// Start 启动设备跟踪
// 通过执行 "adb track-devices" 命令实时监听设备变化
func (t *DeviceTracker) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return fmt.Errorf("设备跟踪器已在运行")
	}
	t.running = true
	t.mu.Unlock()

	ctx, cancelFunc := context.WithCancel(ctx)
	t.cancelFunc = cancelFunc

	logInfo("启动设备跟踪...")

	// 启动 adb track-devices 命令
	cmd := exec.CommandContext(ctx, t.client.ExecPath, "track-devices", "-l")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancelFunc()
		return fmt.Errorf("创建输出管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancelFunc()
		return fmt.Errorf("启动设备跟踪失败: %w", err)
	}

	// 已知设备列表，用于检测变化
	knownDevices := make(map[string]types.Device)

	// 先获取当前已连接设备
	devices, err := t.client.ListDevices(ctx)
	if err == nil {
		for _, d := range devices {
			knownDevices[d.Serial] = d
			if t.onAdd != nil {
				t.onAdd(d)
			}
		}
	}

	// 在 goroutine 中读取跟踪输出
	go func() {
		defer func() {
			t.mu.Lock()
			t.running = false
			t.mu.Unlock()
			cancelFunc()
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			logDebug("设备跟踪输出: %s", line)

			// 解析新的设备列表
			newDevices := make(map[string]types.Device)
			lines := strings.Split(line, "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if l == "" || strings.HasPrefix(l, "List of") {
					continue
				}
				device := parseDeviceLine(l)
				if device.Serial != "" {
					newDevices[device.Serial] = device
				}
			}

			// 检测新增设备
			t.mu.RLock()
			onAdd := t.onAdd
			onRemove := t.onRemove
			onChange := t.onChange
			t.mu.RUnlock()

			for serial, device := range newDevices {
				if _, exists := knownDevices[serial]; !exists {
					logInfo("设备上线: %s", serial)
					if onAdd != nil {
						onAdd(device)
					}
				} else {
					// 设备状态可能变化
					if onChange != nil {
						onChange(device)
					}
				}
			}

			// 检测移除设备
			for serial, device := range knownDevices {
				if _, exists := newDevices[serial]; !exists {
					logInfo("设备离线: %s", serial)
					if onRemove != nil {
						onRemove(device)
					}
				}
			}

			// 更新已知设备列表
			knownDevices = newDevices
		}

		if err := cmd.Wait(); err != nil {
			logError("设备跟踪命令退出: %v", err)
		}
	}()

	return nil
}

// Stop 停止设备跟踪
func (t *DeviceTracker) Stop() {
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
}

// IsRunning 检查跟踪器是否正在运行
func (t *DeviceTracker) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.running
}
