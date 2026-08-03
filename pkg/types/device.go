// Package types 定义 Go-scrcpy 库的公共类型
package types

// Device 表示一个通过 ADB 连接的 Android 设备
type Device struct {
	Serial    string // 设备序列号
	State     string // 设备状态: "device", "offline", "bootloader"
	Model     string // 设备型号 (如 "Pixel 6")
	Product   string // 产品名称
	Transport string // 传输方式 (如 "usb:", "tcp:192.168.1.100:5555")
}

// DeviceEvent 表示设备上下线事件
type DeviceEvent struct {
	Type  EventType // 事件类型
	Device Device   // 关联的设备信息
}

// EventType 设备事件类型
type EventType int

const (
	// EventDeviceAdded 设备上线
	EventDeviceAdded EventType = iota
	// EventDeviceRemoved 设备离线
	EventDeviceRemoved
	// EventDeviceChanged 设备状态变化
	EventDeviceChanged
)

// DeviceInfo 表示设备详细信息（用于 session 级别）
type DeviceInfo struct {
	Serial        string            // 设备序列号
	Model         string            // 设备型号
	AndroidVersion int              // Android 版本号
	SDKVersion    int               // SDK 版本号
	Properties    map[string]string // 设备属性
}

// DisplayInfo 表示设备显示器信息（握手阶段解析）
// 协议格式: 6 × int32 BE = 24 字节
type DisplayInfo struct {
	DisplayID  int32 // 显示器 ID (0 = DEFAULT_DISPLAY)
	Width      int32 // 宽度 (像素)
	Height     int32 // 高度 (像素)
	Rotation   int32 // 旋转 (0, 1, 2, 3)
	LayerStack int32 // 图层栈
	Flags      int32 // 标志位
}

// 显示器标志位常量
const (
	FlagSupportsProtectedBuffers int32 = 0x001 // 支持受保护缓冲区
	FlagSecure                   int32 = 0x002 // 安全显示
	FlagPrivate                  int32 = 0x004 // 私有显示
	FlagPresentation             int32 = 0x008 // 演示显示
	FlagRound                    int32 = 0x010 // 圆形显示
)

// ScreenInfo 表示屏幕信息（25 字节）
type ScreenInfo struct {
	ContentRect   Rect  // 内容区域
	VideoSize     Size  // 视频尺寸
	DeviceRotation uint8 // 设备旋转
}

// Rect 表示矩形区域
type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// Size 表示尺寸
type Size struct {
	Width  int32
	Height int32
}

// VideoFrame 表示一帧解码后的视频数据
type VideoFrame struct {
	PTS    int64  // 显示时间戳 (微秒)
	Data   []byte // 原始编码数据
	Width  int    // 帧宽度
	Height int    // 帧高度
	Config bool   // 是否为配置帧 (SPS/PPS)
	Key    bool   // 是否为关键帧
}

// AudioPacket 表示一个编码后的音频包
type AudioPacket struct {
	PTS  int64  // 显示时间戳 (微秒)
	Data []byte // 编码数据
}
