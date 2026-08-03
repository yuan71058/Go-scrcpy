package scrcpy

import (
	"context"
	"fmt"
	"sync"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/types"
	"github.com/yuan71058/go-scrcpy/pkg/video"
)

// MultiClient 多设备管理器
// 支持多个设备同时接入控制
type MultiClient struct {
	clients map[string]*Client // keyed by serial
	adb     *adb.Client
	mu      sync.RWMutex
	closed  bool
}

// NewMulti 创建多设备管理器
func NewMulti(adbClient *adb.Client) *MultiClient {
	return &MultiClient{
		clients: make(map[string]*Client),
		adb:     adbClient,
	}
}

// Add 添加设备
// serial: 设备序列号
// opts: 启动选项
func (m *MultiClient) Add(serial string, opts Options) (*Client, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, fmt.Errorf("管理器已关闭")
	}

	// 检查是否已存在
	if _, exists := m.clients[serial]; exists {
		m.mu.Unlock()
		return nil, fmt.Errorf("设备 %s 已存在", serial)
	}
	m.mu.Unlock()

	logInfo("添加设备 [%s]", serial)

	// 创建客户端
	client := New(serial, opts)

	// 启动客户端
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("启动设备失败: %w", err)
	}

	// 添加到管理器
	m.mu.Lock()
	m.clients[serial] = client
	m.mu.Unlock()

	logInfo("设备添加成功 [%s]", serial)
	return client, nil
}

// Remove 移除设备
func (m *MultiClient) Remove(serial string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[serial]
	if !exists {
		return fmt.Errorf("设备 %s 不存在", serial)
	}

	logInfo("移除设备 [%s]", serial)

	// 关闭客户端
	if err := client.Close(); err != nil {
		logError("关闭设备失败 [%s]: %v", serial, err)
	}

	delete(m.clients, serial)
	logInfo("设备已移除 [%s]", serial)
	return nil
}

// Get 获取指定设备的客户端
func (m *MultiClient) Get(serial string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serial]
	return client, exists
}

// List 列出所有设备
func (m *MultiClient) List() []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var clients []*Client
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients
}

// Count 获取设备数量
func (m *MultiClient) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// Broadcast 向所有设备发送控制消息
func (m *MultiClient) Broadcast(msg []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for serial, client := range m.clients {
		if err := client.SendControl(msg); err != nil {
			logError("广播到设备 [%s] 失败: %v", serial, err)
			continue
		}
	}
	return nil
}

// ForEach 遍历所有设备执行操作
func (m *MultiClient) ForEach(fn func(serial string, client *Client)) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for serial, client := range m.clients {
		fn(serial, client)
	}
}

// VideoStreamAll 获取所有设备的视频帧通道
// 返回一个合并的通道
func (m *MultiClient) VideoStreamAll() <-chan *VideoFrameWithSerial {
	ch := make(chan *VideoFrameWithSerial, 100)

	m.mu.RLock()
	clients := make(map[string]*Client)
	for k, v := range m.clients {
		clients[k] = v
	}
	m.mu.RUnlock()

	go func() {
		defer close(ch)

		var wg sync.WaitGroup
		for serial, client := range clients {
			wg.Add(1)
			go func(s string, c *Client) {
				defer wg.Done()
				for frame := range c.VideoStream() {
					select {
					case ch <- &VideoFrameWithSerial{
						Serial: s,
						Frame:  frame,
					}:
					default:
						// 通道满，丢弃
					}
				}
			}(serial, client)
		}
		wg.Wait()
	}()

	return ch
}

// VideoFrameWithSerial 带设备序列号的视频帧
type VideoFrameWithSerial struct {
	Serial string
	Frame  *video.DecodedFrame
}

// GetDevices 获取已连接的设备列表
func GetDevices(adbClient *adb.Client) ([]types.Device, error) {
	ctx := context.Background()
	return adbClient.ListDevices(ctx)
}

// WatchDevices 监听设备上下线
// 返回添加和移除回调函数
func WatchDevices(adbClient *adb.Client, onAdd func(serial string), onRemove func(serial string)) *adb.DeviceTracker {
	tracker := adb.NewDeviceTracker(adbClient)

	tracker.OnAdd(func(device types.Device) {
		if onAdd != nil {
			onAdd(device.Serial)
		}
	})

	tracker.OnRemove(func(device types.Device) {
		if onRemove != nil {
			onRemove(device.Serial)
		}
	})

	return tracker
}

// Close 关闭管理器
func (m *MultiClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	logInfo("关闭多设备管理器")

	var firstErr error
	for serial, client := range m.clients {
		if err := client.Close(); err != nil {
			logError("关闭设备 [%s] 失败: %v", serial, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	m.clients = make(map[string]*Client)
	return firstErr
}
