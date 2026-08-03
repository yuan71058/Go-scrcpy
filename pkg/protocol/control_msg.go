package protocol

import (
	"encoding/binary"
	"fmt"
)

// ControlMessageBuilder 控制消息构建器
// 用于构建发送给 scrcpy-server 的控制消息
type ControlMessageBuilder struct {
	buf []byte
}

// NewControlMessageBuilder 创建控制消息构建器
func NewControlMessageBuilder() *ControlMessageBuilder {
	return &ControlMessageBuilder{
		buf: make([]byte, 0, 256),
	}
}

// Build 构建消息并返回字节切片
func (b *ControlMessageBuilder) Build() []byte {
	result := make([]byte, len(b.buf))
	copy(result, b.buf)
	return result
}

// InjectKeycode 构建按键注入消息 (type=0, 14 字节)
// action: 0=DOWN, 1=UP, 2=MULTI
// keycode: Android KEYCODE_* 常量
// repeat: 重复次数
// metastate: 修饰键状态 (META_* bitmask)
func InjectKeycode(action byte, keycode int32, repeat int32, metastate int32) []byte {
	buf := make([]byte, 14)
	buf[0] = ControlTypeInjectKeycode
	buf[1] = action
	binary.BigEndian.PutUint32(buf[2:6], uint32(keycode))
	binary.BigEndian.PutUint32(buf[6:10], uint32(repeat))
	binary.BigEndian.PutUint32(buf[10:14], uint32(metastate))
	return buf
}

// InjectText 构建文本注入消息 (type=1)
// text: 要注入的文本 (最长 300 字节)
func InjectText(text string) []byte {
	textBytes := []byte(text)
	if len(textBytes) > MaxTextLength {
		textBytes = textBytes[:MaxTextLength]
	}

	buf := make([]byte, 5+len(textBytes))
	buf[0] = ControlTypeInjectText
	binary.BigEndian.PutUint32(buf[1:5], uint32(len(textBytes)))
	copy(buf[5:], textBytes)
	return buf
}

// InjectTouchEvent 构建触摸事件消息 (type=2, 32 字节)
// action: 0=DOWN, 1=UP, 2=MOVE, 3=CANCEL
// pointerID: 指针 ID (UINT64_MAX=mouse, UINT64_MAX-1=generic finger)
// x, y: 触摸坐标
// screenW, screenH: 屏幕尺寸
// pressure: 压力值 (0x0000=0.0, 0xFFFF=1.0)
// actionButton: 触摸按钮
// buttons: 按钮状态
func InjectTouchEvent(action byte, pointerID uint64, x, y int32, screenW, screenH uint16, pressure uint16, actionButton, buttons int32) []byte {
	buf := make([]byte, 32)
	buf[0] = ControlTypeInjectTouchEvent
	buf[1] = action

	// pointer ID (big-endian uint64)
	buf[2] = byte(pointerID >> 56)
	buf[3] = byte(pointerID >> 48)
	buf[4] = byte(pointerID >> 40)
	buf[5] = byte(pointerID >> 32)
	buf[6] = byte(pointerID >> 24)
	buf[7] = byte(pointerID >> 16)
	buf[8] = byte(pointerID >> 8)
	buf[9] = byte(pointerID)

	// x, y (big-endian int32)
	binary.BigEndian.PutUint32(buf[10:14], uint32(x))
	binary.BigEndian.PutUint32(buf[14:18], uint32(y))

	// screenW, screenH (big-endian uint16)
	binary.BigEndian.PutUint16(buf[18:20], screenW)
	binary.BigEndian.PutUint16(buf[20:22], screenH)

	// pressure (big-endian uint16)
	binary.BigEndian.PutUint16(buf[22:24], pressure)

	// actionButton, buttons (big-endian int32)
	binary.BigEndian.PutUint32(buf[24:28], uint32(actionButton))
	binary.BigEndian.PutUint32(buf[28:32], uint32(buttons))

	return buf
}

// InjectScrollEvent 构建滚动事件消息 (type=3, 21 字节)
// x, y: 滚动位置
// screenW, screenH: 屏幕尺寸
// hscroll: 水平滚动量 (乘以 16 的定点数)
// vscroll: 垂直滚动量 (乘以 16 的定点数)
// buttons: 按钮状态
func InjectScrollEvent(x, y int32, screenW, screenH uint16, hscroll, vscroll int16, buttons int32) []byte {
	buf := make([]byte, 21)
	buf[0] = ControlTypeInjectScrollEvent

	// x, y
	binary.BigEndian.PutUint32(buf[1:5], uint32(x))
	binary.BigEndian.PutUint32(buf[5:9], uint32(y))

	// screenW, screenH
	binary.BigEndian.PutUint16(buf[9:11], screenW)
	binary.BigEndian.PutUint16(buf[11:13], screenH)

	// hscroll, vscroll
	binary.BigEndian.PutUint16(buf[13:15], uint16(hscroll))
	binary.BigEndian.PutUint16(buf[15:17], uint16(vscroll))

	// buttons
	binary.BigEndian.PutUint32(buf[17:21], uint32(buttons))

	return buf
}

// BackOrScreenOn 构建返回键/亮屏消息 (type=4, 2 字节)
// action: 0=DOWN, 1=UP
func BackOrScreenOn(action byte) []byte {
	return []byte{ControlTypeBackOrScreenOn, action}
}

// ExpandNotificationPanel 构建展开通知栏消息 (type=5)
func ExpandNotificationPanel() []byte {
	return []byte{ControlTypeExpandNotificationPanel}
}

// ExpandSettingsPanel 构建展开设置面板消息 (type=6)
func ExpandSettingsPanel() []byte {
	return []byte{ControlTypeExpandSettingsPanel}
}

// CollapsePanels 构建收起面板消息 (type=7)
func CollapsePanels() []byte {
	return []byte{ControlTypeCollapsePanels}
}

// GetClipboard 构建获取剪贴板消息 (type=8)
// copyKey: 0=NONE, 1=COPY, 2=CUT
func GetClipboard(copyKey byte) []byte {
	return []byte{ControlTypeGetClipboard, copyKey}
}

// SetClipboard 构建设置剪贴板消息 (type=9)
// text: 剪贴板内容
// paste: 是否自动粘贴
// sequence: 序列号 (单调递增)
func SetClipboard(text string, paste bool, sequence uint64) []byte {
	textBytes := []byte(text)
	if len(textBytes) > MaxClipboardTextSize {
		textBytes = textBytes[:MaxClipboardTextSize]
	}

	buf := make([]byte, 10+len(textBytes))
	buf[0] = ControlTypeSetClipboard

	// sequence (big-endian uint64)
	buf[1] = byte(sequence >> 56)
	buf[2] = byte(sequence >> 48)
	buf[3] = byte(sequence >> 40)
	buf[4] = byte(sequence >> 32)
	buf[5] = byte(sequence >> 24)
	buf[6] = byte(sequence >> 16)
	buf[7] = byte(sequence >> 8)
	buf[8] = byte(sequence)

	// paste
	if paste {
		buf[9] = 1
	}

	// text length + text
	binary.BigEndian.PutUint32(buf[10:14], uint32(len(textBytes)))
	copy(buf[14:], textBytes)

	return buf
}

// SetDisplayPower 构建设置屏幕电源消息 (type=10)
// on: true=亮屏, false=息屏
func SetDisplayPower(on bool) []byte {
	if on {
		return []byte{ControlTypeSetDisplayPower, 1}
	}
	return []byte{ControlTypeSetDisplayPower, 0}
}

// RotateDevice 构建设备旋转消息 (type=11)
func RotateDevice() []byte {
	return []byte{ControlTypeRotateDevice}
}

// StartApp 构建启动应用消息 (type=16)
// packageName: 应用包名
func StartApp(packageName string) []byte {
	nameBytes := []byte(packageName)
	if len(nameBytes) > MaxFileNameLength {
		nameBytes = nameBytes[:MaxFileNameLength]
	}

	buf := make([]byte, 2+len(nameBytes))
	buf[0] = ControlTypeStartApp
	buf[1] = byte(len(nameBytes))
	copy(buf[2:], nameBytes)

	return buf
}

// ResetVideo 构建重置视频消息 (type=17)
func ResetVideo() []byte {
	return []byte{ControlTypeResetVideo}
}

// CameraSetTorch 构建设置相机闪光灯消息 (type=18)
// on: true=开启, false=关闭
func CameraSetTorch(on bool) []byte {
	if on {
		return []byte{ControlTypeCameraSetTorch, 1}
	}
	return []byte{ControlTypeCameraSetTorch, 0}
}

// CameraZoomIn 构建相机放大消息 (type=19)
func CameraZoomIn() []byte {
	return []byte{ControlTypeCameraZoomIn}
}

// CameraZoomOut 构建相机缩小消息 (type=20)
func CameraZoomOut() []byte {
	return []byte{ControlTypeCameraZoomOut}
}

// ResizeDisplay 构建调整显示器大小消息 (type=21)
// width, height: 新的显示器尺寸
func ResizeDisplay(width, height uint16) []byte {
	buf := make([]byte, 5)
	buf[0] = ControlTypeResizeDisplay
	binary.BigEndian.PutUint16(buf[1:3], width)
	binary.BigEndian.PutUint16(buf[3:5], height)
	return buf
}

// ScanFile 构建扫描文件消息 (type=22)
// path: 文件路径
func ScanFile(path string) []byte {
	pathBytes := []byte(path)
	if len(pathBytes) > MaxFilePathLength {
		pathBytes = pathBytes[:MaxFilePathLength]
	}

	buf := make([]byte, 5+len(pathBytes))
	buf[0] = ControlTypeScanFile
	binary.BigEndian.PutUint32(buf[1:5], uint32(len(pathBytes)))
	copy(buf[5:], pathBytes)

	return buf
}

// UHIDCreate 构建创建 UHID 设备消息 (type=12)
// id: 设备 ID
// vendorID: 厂商 ID
// productID: 产品 ID
// name: 设备名称
// reportDesc: HID 报告描述符
func UHIDCreate(id uint16, vendorID, productID uint16, name string, reportDesc []byte) []byte {
	nameBytes := []byte(name)
	if len(nameBytes) > 127 {
		nameBytes = nameBytes[:127]
	}

	buf := make([]byte, 9+len(nameBytes)+2+len(reportDesc))
	buf[0] = ControlTypeUHIDCreate

	// id
	binary.BigEndian.PutUint16(buf[1:3], id)

	// vendor_id
	binary.BigEndian.PutUint16(buf[3:5], vendorID)

	// product_id
	binary.BigEndian.PutUint16(buf[5:7], productID)

	// name_length + name
	buf[7] = byte(len(nameBytes))
	copy(buf[8:8+len(nameBytes)], nameBytes)

	// report_desc_size
	offset := 8 + len(nameBytes)
	binary.BigEndian.PutUint16(buf[offset:offset+2], uint16(len(reportDesc)))
	offset += 2

	// report_desc
	copy(buf[offset:], reportDesc)

	return buf
}

// UHIDInput 构建 UHID 输入消息 (type=13)
// id: 设备 ID
// data: 输入数据
func UHIDInput(id uint16, data []byte) []byte {
	buf := make([]byte, 5+len(data))
	buf[0] = ControlTypeUHIDInput

	// id
	binary.BigEndian.PutUint16(buf[1:3], id)

	// size
	binary.BigEndian.PutUint16(buf[3:5], uint16(len(data)))

	// data
	copy(buf[5:], data)

	return buf
}

// UHIDDestroy 构建销毁 UHID 设备消息 (type=14)
// id: 设备 ID
func UHIDDestroy(id uint16) []byte {
	buf := make([]byte, 3)
	buf[0] = ControlTypeUHIDDestroy

	// id
	binary.BigEndian.PutUint16(buf[1:3], id)

	return buf
}

// ControlMessageToString 控制消息转字符串（用于调试）
func ControlMessageToString(msg []byte) string {
	if len(msg) == 0 {
		return "EMPTY"
	}

	switch msg[0] {
	case ControlTypeInjectKeycode:
		return fmt.Sprintf("INJECT_KEYCODE(action=%d, keycode=%d)", msg[1],
			int32(msg[2])<<24|int32(msg[3])<<16|int32(msg[4])<<8|int32(msg[5]))
	case ControlTypeInjectText:
		return fmt.Sprintf("INJECT_TEXT(length=%d)", len(msg)-5)
	case ControlTypeInjectTouchEvent:
		return fmt.Sprintf("INJECT_TOUCH_EVENT(action=%d)", msg[1])
	case ControlTypeInjectScrollEvent:
		return "INJECT_SCROLL_EVENT"
	case ControlTypeBackOrScreenOn:
		return "BACK_OR_SCREEN_ON"
	case ControlTypeExpandNotificationPanel:
		return "EXPAND_NOTIFICATION_PANEL"
	case ControlTypeExpandSettingsPanel:
		return "EXPAND_SETTINGS_PANEL"
	case ControlTypeCollapsePanels:
		return "COLLAPSE_PANELS"
	case ControlTypeGetClipboard:
		return "GET_CLIPBOARD"
	case ControlTypeSetClipboard:
		return "SET_CLIPBOARD"
	case ControlTypeSetDisplayPower:
		return "SET_DISPLAY_POWER"
	case ControlTypeRotateDevice:
		return "ROTATE_DEVICE"
	case ControlTypeStartApp:
		return "START_APP"
	case ControlTypeResetVideo:
		return "RESET_VIDEO"
	case ControlTypeCameraSetTorch:
		return "CAMERA_SET_TORCH"
	case ControlTypeCameraZoomIn:
		return "CAMERA_ZOOM_IN"
	case ControlTypeCameraZoomOut:
		return "CAMERA_ZOOM_OUT"
	case ControlTypeResizeDisplay:
		return "RESIZE_DISPLAY"
	case ControlTypeScanFile:
		return "SCAN_FILE"
	case ControlTypeUHIDCreate:
		return "UHID_CREATE"
	case ControlTypeUHIDInput:
		return "UHID_INPUT"
	case ControlTypeUHIDDestroy:
		return "UHID_DESTROY"
	default:
		return fmt.Sprintf("UNKNOWN(type=%d)", msg[0])
	}
}
