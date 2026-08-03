package input

import (
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
)

// Android KeyEvent 操作常量
const (
	KeyActionDown = protocol.KeyActionDown
	KeyActionUp   = protocol.KeyActionUp
	KeyActionMulti = protocol.KeyActionMulti
)

// Android 修饰键状态 (META_* bitmask)
const (
	MetaNone         int32 = 0
	MetaAltLeft      int32 = 0x10    // 左 Alt
	MetaAltRight     int32 = 0x20    // 右 Alt
	MetaShiftLeft    int32 = 0x40    // 左 Shift
	MetaShiftRight   int32 = 0x80    // 右 Shift
	MetaCtrlLeft     int32 = 0x1000  // 左 Ctrl
	MetaCtrlRight    int32 = 0x2000  // 右 Ctrl
	MetaMetaLeft     int32 = 0x10000 // 左 Meta/Win
	MetaMetaRight    int32 = 0x20000 // 右 Meta/Win
	MetaCapsLock     int32 = 0x100000 // CapsLock
	MetaNumLock      int32 = 0x200000 // NumLock
	MetaScrollLock   int32 = 0x400000 // ScrollLock
)

// KeyDown 构建按键按下事件
// keycode: Android KEYCODE_* 常量
// meta: 修饰键状态 (MetaShiftLeft | MetaCtrlLeft 等)
func KeyDown(keycode int32, meta int32) []byte {
	return protocol.InjectKeycode(KeyActionDown, keycode, 0, meta)
}

// KeyUp 构建按键抬起事件
func KeyUp(keycode int32, meta int32) []byte {
	return protocol.InjectKeycode(KeyActionUp, keycode, 0, meta)
}

// KeyPress 构建完整的按键事件 (按下 + 抬起)
func KeyPress(keycode int32, meta int32) [][]byte {
	return [][]byte{
		KeyDown(keycode, meta),
		KeyUp(keycode, meta),
	}
}

// Text 构建文本注入事件
// text: 要注入的文本 (最长 300 字节)
func Text(text string) []byte {
	return protocol.InjectText(text)
}

// BackOrScreenOn 构建返回键/亮屏事件
// action: 0=DOWN, 1=UP
func BackOrScreenOn(action byte) []byte {
	return protocol.BackOrScreenOn(action)
}

// BackPress 构建返回键完整事件 (按下 + 抬起)
func BackPress() [][]byte {
	return [][]byte{
		BackOrScreenOn(KeyActionDown),
		BackOrScreenOn(KeyActionUp),
	}
}

// ExpandNotificationPanel 构建展开通知栏事件
func ExpandNotificationPanel() []byte {
	return protocol.ExpandNotificationPanel()
}

// ExpandSettingsPanel 构建展开设置面板事件
func ExpandSettingsPanel() []byte {
	return protocol.ExpandSettingsPanel()
}

// CollapsePanels 构建收起面板事件
func CollapsePanels() []byte {
	return protocol.CollapsePanels()
}

// RotateDevice 构建设备旋转事件
func RotateDevice() []byte {
	return protocol.RotateDevice()
}

// SetDisplayPower 构建设置屏幕电源事件
// on: true=亮屏, false=息屏
func SetDisplayPower(on bool) []byte {
	return protocol.SetDisplayPower(on)
}

// PowerOn 构建亮屏事件
func PowerOn() []byte {
	return SetDisplayPower(true)
}

// PowerOff 构建息屏事件
func PowerOff() []byte {
	return SetDisplayPower(false)
}

// StartApp 构建启动应用事件
// packageName: 应用包名
func StartApp(packageName string) []byte {
	return protocol.StartApp(packageName)
}

// ResetVideo 构建重置视频事件
func ResetVideo() []byte {
	return protocol.ResetVideo()
}

// CameraSetTorch 构建设置相机闪光灯事件
// on: true=开启, false=关闭
func CameraSetTorch(on bool) []byte {
	return protocol.CameraSetTorch(on)
}

// CameraZoomIn 构建相机放大事件
func CameraZoomIn() []byte {
	return protocol.CameraZoomIn()
}

// CameraZoomOut 构建相机缩小事件
func CameraZoomOut() []byte {
	return protocol.CameraZoomOut()
}

// ResizeDisplay 构建调整显示器大小事件
// width, height: 新的显示器尺寸
func ResizeDisplay(width, height uint16) []byte {
	return protocol.ResizeDisplay(width, height)
}

// ScanFile 构建扫描文件事件
// path: 文件路径
func ScanFile(path string) []byte {
	return protocol.ScanFile(path)
}
