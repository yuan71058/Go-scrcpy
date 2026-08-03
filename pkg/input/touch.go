// Package input 提供输入事件构建函数
// 所有函数返回可直接通过控制通道发送的字节切片
package input

import (
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
)

// TouchDown 构建触摸按下事件
// pointerID: 指针 ID (0xFFFFFFFFFFFFFFFF=鼠标, 0xFFFFFFFFFFFFFFFE=通用手指)
// x, y: 触摸坐标
// screenW, screenH: 屏幕尺寸
func TouchDown(pointerID uint64, x, y int32, screenW, screenH uint16) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionDown,
		pointerID,
		x, y,
		screenW, screenH,
		0xFFFF, // pressure = 1.0
		0,      // actionButton
		0,      // buttons
	)
}

// TouchMove 构建触摸移动事件
func TouchMove(pointerID uint64, x, y int32, screenW, screenH uint16) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionMove,
		pointerID,
		x, y,
		screenW, screenH,
		0xFFFF, // pressure = 1.0
		0,      // actionButton
		0,      // buttons
	)
}

// TouchUp 构建触摸抬起事件
func TouchUp(pointerID uint64, x, y int32, screenW, screenH uint16) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionUp,
		pointerID,
		x, y,
		screenW, screenH,
		0x0000, // pressure = 0.0
		0,      // actionButton
		0,      // buttons
	)
}

// TouchCancel 构建触摸取消事件
func TouchCancel(pointerID uint64, x, y int32, screenW, screenH uint16) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionCancel,
		pointerID,
		x, y,
		screenW, screenH,
		0x0000, // pressure = 0.0
		0,      // actionButton
		0,      // buttons
	)
}

// MouseDown 构建鼠标按下事件
// button: 1=左键, 2=右键, 4=中键
func MouseDown(x, y int32, screenW, screenH uint16, button int32) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionDown,
		protocol.PointerIDMouse,
		x, y,
		screenW, screenH,
		0xFFFF,   // pressure = 1.0
		button,   // actionButton
		button,   // buttons
	)
}

// MouseMove 构建鼠标移动事件
func MouseMove(x, y int32, screenW, screenH uint16, buttons int32) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionMove,
		protocol.PointerIDMouse,
		x, y,
		screenW, screenH,
		0xFFFF,   // pressure = 1.0
		0,        // actionButton
		buttons,  // buttons
	)
}

// MouseUp 构建鼠标抬起事件
func MouseUp(x, y int32, screenW, screenH uint16, button int32) []byte {
	return protocol.InjectTouchEvent(
		protocol.TouchActionUp,
		protocol.PointerIDMouse,
		x, y,
		screenW, screenH,
		0x0000, // pressure = 0.0
		button, // actionButton
		0,      // buttons
	)
}

// Scroll 构建滚动事件
// x, y: 滚动位置
// screenW, screenH: 屏幕尺寸
// hscroll: 水平滚动量 (正数向右, 负数向左)
// vscroll: 垂直滚动量 (正数向下, 负数向上)
func Scroll(x, y int32, screenW, screenH uint16, hscroll, vscroll int16) []byte {
	return protocol.InjectScrollEvent(x, y, screenW, screenH, hscroll, vscroll, 0)
}
