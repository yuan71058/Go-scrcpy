package scrcpy

import (
	"context"
	"fmt"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/audio"
	"github.com/yuan71058/go-scrcpy/pkg/input"
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
	"github.com/yuan71058/go-scrcpy/pkg/video"
)

// Client 单设备客户端
// 封装与单个 Android 设备的 scrcpy 连接
type Client struct {
	session *DeviceSession
	adb     *adb.Client
	opts    Options
	closed  bool
}

// New 创建新的单设备客户端
// serial: 设备序列号
// opts: 启动选项
func New(serial string, opts Options) *Client {
	return &Client{
		session: NewSession(serial),
		adb:     adb.NewClient(opts.ADBPath),
		opts:    opts,
	}
}

// Start 启动客户端
// 建立连接并开始接收视频流
func (c *Client) Start(ctx context.Context) error {
	if c.closed {
		return fmt.Errorf("客户端已关闭")
	}

	logInfo("启动客户端 [%s]", c.session.GetSerial())

	// 连接设备
	if err := c.session.Connect(ctx, c.adb, c.opts.Server, c.opts.LocalJAR); err != nil {
		return fmt.Errorf("连接设备失败: %w", err)
	}

	// 启动会话
	if err := c.session.Start(ctx); err != nil {
		return fmt.Errorf("启动会话失败: %w", err)
	}

	logInfo("客户端启动成功 [%s]", c.session.GetSerial())
	return nil
}

// VideoStream 返回视频帧通道
func (c *Client) VideoStream() <-chan *video.DecodedFrame {
	return c.session.VideoStream()
}

// AudioStream 返回音频数据通道
func (c *Client) AudioStream() <-chan *audio.AudioData {
	return c.session.AudioStream()
}

// SendControl 发送控制消息
func (c *Client) SendControl(msg []byte) error {
	return c.session.SendControl(msg)
}

// DeviceInfo 获取设备信息
func (c *Client) DeviceInfo() interface{} {
	return c.session.GetDeviceInfo()
}

// Handshake 获取握手数据
func (c *Client) Handshake() interface{} {
	return c.session.GetHandshake()
}

// Serial 获取设备序列号
func (c *Client) Serial() string {
	return c.session.GetSerial()
}

// === 便捷输入方法 ===

// TouchDown 触摸按下
func (c *Client) TouchDown(pointerID uint64, x, y int32, screenW, screenH uint16) error {
	msg := input.TouchDown(pointerID, x, y, screenW, screenH)
	return c.SendControl(msg)
}

// TouchMove 触摸移动
func (c *Client) TouchMove(pointerID uint64, x, y int32, screenW, screenH uint16) error {
	msg := input.TouchMove(pointerID, x, y, screenW, screenH)
	return c.SendControl(msg)
}

// TouchUp 触摸抬起
func (c *Client) TouchUp(pointerID uint64, x, y int32, screenW, screenH uint16) error {
	msg := input.TouchUp(pointerID, x, y, screenW, screenH)
	return c.SendControl(msg)
}

// MouseDown 鼠标按下
func (c *Client) MouseDown(x, y int32, screenW, screenH uint16, button int32) error {
	msg := input.MouseDown(x, y, screenW, screenH, button)
	return c.SendControl(msg)
}

// MouseMove 鼠标移动
func (c *Client) MouseMove(x, y int32, screenW, screenH uint16, buttons int32) error {
	msg := input.MouseMove(x, y, screenW, screenH, buttons)
	return c.SendControl(msg)
}

// MouseUp 鼠标抬起
func (c *Client) MouseUp(x, y int32, screenW, screenH uint16, button int32) error {
	msg := input.MouseUp(x, y, screenW, screenH, button)
	return c.SendControl(msg)
}

// Scroll 滚动
func (c *Client) Scroll(x, y int32, screenW, screenH uint16, hscroll, vscroll int16) error {
	msg := input.Scroll(x, y, screenW, screenH, hscroll, vscroll)
	return c.SendControl(msg)
}

// KeyDown 按键按下
func (c *Client) KeyDown(keycode int32, meta int32) error {
	msg := input.KeyDown(keycode, meta)
	return c.SendControl(msg)
}

// KeyUp 按键抬起
func (c *Client) KeyUp(keycode int32, meta int32) error {
	msg := input.KeyUp(keycode, meta)
	return c.SendControl(msg)
}

// KeyPress 完整按键事件 (按下 + 抬起)
func (c *Client) KeyPress(keycode int32, meta int32) error {
	msgs := input.KeyPress(keycode, meta)
	for _, msg := range msgs {
		if err := c.SendControl(msg); err != nil {
			return err
		}
	}
	return nil
}

// Text 文本注入
func (c *Client) Text(text string) error {
	msg := input.Text(text)
	return c.SendControl(msg)
}

// Back 返回键
func (c *Client) Back() error {
	msgs := input.BackPress()
	for _, msg := range msgs {
		if err := c.SendControl(msg); err != nil {
			return err
		}
	}
	return nil
}

// Home Home 键
func (c *Client) Home() error {
	return c.KeyPress(input.KeycodeHome, input.MetaNone)
}

// Power 电源键
func (c *Client) Power() error {
	return c.KeyPress(input.KeycodePower, input.MetaNone)
}

// VolumeUp 音量加
func (c *Client) VolumeUp() error {
	return c.KeyPress(input.KeycodeVolumeUp, input.MetaNone)
}

// VolumeDown 音量减
func (c *Client) VolumeDown() error {
	return c.KeyPress(input.KeycodeVolumeDown, input.MetaNone)
}

// Menu 菜单键
func (c *Client) Menu() error {
	return c.KeyPress(input.KeycodeMenu, input.MetaNone)
}

// Enter 回车键
func (c *Client) Enter() error {
	return c.KeyPress(input.KeycodeEnter, input.MetaNone)
}

// Delete 删除键
func (c *Client) Delete() error {
	return c.KeyPress(input.KeycodeDel, input.MetaNone)
}

// Tab Tab 键
func (c *Client) Tab() error {
	return c.KeyPress(input.KeycodeTab, input.MetaNone)
}

// Escape ESC 键
func (c *Client) Escape() error {
	return c.KeyPress(input.KeycodeEscape, input.MetaNone)
}

// Space 空格键
func (c *Client) Space() error {
	return c.KeyPress(input.KeycodeSpace, input.MetaNone)
}

// === 面板控制 ===

// ExpandNotificationPanel 展开通知栏
func (c *Client) ExpandNotificationPanel() error {
	msg := input.ExpandNotificationPanel()
	return c.SendControl(msg)
}

// ExpandSettingsPanel 展开设置面板
func (c *Client) ExpandSettingsPanel() error {
	msg := input.ExpandSettingsPanel()
	return c.SendControl(msg)
}

// CollapsePanels 收起面板
func (c *Client) CollapsePanels() error {
	msg := input.CollapsePanels()
	return c.SendControl(msg)
}

// RotateDevice 旋转设备
func (c *Client) RotateDevice() error {
	msg := input.RotateDevice()
	return c.SendControl(msg)
}

// SetDisplayPower 设置屏幕电源
func (c *Client) SetDisplayPower(on bool) error {
	msg := input.SetDisplayPower(on)
	return c.SendControl(msg)
}

// PowerOn 亮屏
func (c *Client) PowerOn() error {
	return c.SetDisplayPower(true)
}

// PowerOff 息屏
func (c *Client) PowerOff() error {
	return c.SetDisplayPower(false)
}

// === 剪贴板 ===

// SetClipboard 设置剪贴板
func (c *Client) SetClipboard(text string, paste bool, sequence uint64) error {
	msg := protocol.SetClipboard(text, paste, sequence)
	return c.SendControl(msg)
}

// GetClipboard 获取剪贴板
func (c *Client) GetClipboard() error {
	msg := protocol.GetClipboard(0)
	return c.SendControl(msg)
}

// === 应用控制 ===

// StartApp 启动应用
func (c *Client) StartApp(packageName string) error {
	msg := input.StartApp(packageName)
	return c.SendControl(msg)
}

// ScanFile 扫描文件
func (c *Client) ScanFile(path string) error {
	msg := input.ScanFile(path)
	return c.SendControl(msg)
}

// === 显示控制 ===

// ResizeDisplay 调整显示器大小
func (c *Client) ResizeDisplay(width, height uint16) error {
	msg := input.ResizeDisplay(width, height)
	return c.SendControl(msg)
}

// === 摄像头控制 ===

// CameraSetTorch 设置相机闪光灯
func (c *Client) CameraSetTorch(on bool) error {
	msg := input.CameraSetTorch(on)
	return c.SendControl(msg)
}

// CameraZoomIn 相机放大
func (c *Client) CameraZoomIn() error {
	msg := input.CameraZoomIn()
	return c.SendControl(msg)
}

// CameraZoomOut 相机缩小
func (c *Client) CameraZoomOut() error {
	msg := input.CameraZoomOut()
	return c.SendControl(msg)
}

// === UHID 设备 ===

// UHIDCreate 创建 UHID 设备
func (c *Client) UHIDCreate(id uint16, vendorID, productID uint16, name string, reportDesc []byte) error {
	msg := protocol.UHIDCreate(id, vendorID, productID, name, reportDesc)
	return c.SendControl(msg)
}

// UHIDInput UHID 输入
func (c *Client) UHIDInput(id uint16, data []byte) error {
	msg := protocol.UHIDInput(id, data)
	return c.SendControl(msg)
}

// UHIDDestroy 销毁 UHID 设备
func (c *Client) UHIDDestroy(id uint16) error {
	msg := protocol.UHIDDestroy(id)
	return c.SendControl(msg)
}

// === 重置 ===

// ResetVideo 重置视频
func (c *Client) ResetVideo() error {
	msg := input.ResetVideo()
	return c.SendControl(msg)
}

// Close 关闭客户端
func (c *Client) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	logInfo("关闭客户端 [%s]", c.session.GetSerial())

	return c.session.Close()
}
