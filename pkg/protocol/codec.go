package protocol

// 视频编解码器 ID (4 字节 ASCII/BE int32)
// Server 在视频流头部发送这些 ID
const (
	// VideoCodecIDH264 H.264/AVC 编解码器 ID
	VideoCodecIDH264 uint32 = 0x68323634 // ASCII: "h264"
	// VideoCodecIDH265 H.265/HEVC 编解码器 ID
	VideoCodecIDH265 uint32 = 0x68323635 // ASCII: "h265"
	// VideoCodecIDAV1 AV1 编解码器 ID
	VideoCodecIDAV1 uint32 = 0x00617631 // ASCII: "\0av1"
	// VideoCodecIDVP8 VP8 编解码器 ID
	VideoCodecIDVP8 uint32 = 0x00767038 // ASCII: "\0vp8"
	// VideoCodecIDVP9 VP9 编解码器 ID
	VideoCodecIDVP9 uint32 = 0x00767039 // ASCII: "\0vp9"
)

// 音频编解码器 ID (4 字节 ASCII/BE int32)
const (
	// AudioCodecIDOpus Opus 编解码器 ID
	AudioCodecIDOpus uint32 = 0x6F707573 // ASCII: "opus"
	// AudioCodecIDAAC AAC 编解码器 ID
	AudioCodecIDAAC uint32 = 0x00616163 // ASCII: "\0aac"
	// AudioCodecIDFLAC FLAC 编解码器 ID
	AudioCodecIDFLAC uint32 = 0x666C6163 // ASCII: "flac"
	// AudioCodecIDRAW PCM 原始音频编解码器 ID
	AudioCodecIDRAW uint32 = 0x00726177 // ASCII: "\0raw"
)

// 帧标志位常量 (用于 PTS + flags 字段)
// 注意: 这些标志位在协议中使用 uint64 表示
const (
	// PacketFlagConfig 配置包标志 (SPS/PPS) - bit 63
	PacketFlagConfig uint64 = 1 << 63
	// PacketFlagKeyFrame 关键帧标志 - bit 62
	PacketFlagKeyFrame uint64 = 1 << 62
	// PacketPTSMask PTS 掩码 (低 62 位)
	PacketPTSMask uint64 = 0x3FFFFFFFFFFFFFFF
)

// 控制消息类型常量
const (
	// ControlTypeInjectKeycode 注入按键码
	ControlTypeInjectKeycode byte = 0
	// ControlTypeInjectText 注入文本
	ControlTypeInjectText byte = 1
	// ControlTypeInjectTouchEvent 注入触摸事件
	ControlTypeInjectTouchEvent byte = 2
	// ControlTypeInjectScrollEvent 注入滚动事件
	ControlTypeInjectScrollEvent byte = 3
	// ControlTypeBackOrScreenOn 返回键或亮屏
	ControlTypeBackOrScreenOn byte = 4
	// ControlTypeExpandNotificationPanel 展开通知栏
	ControlTypeExpandNotificationPanel byte = 5
	// ControlTypeExpandSettingsPanel 展开设置面板
	ControlTypeExpandSettingsPanel byte = 6
	// ControlTypeCollapsePanels 收起所有面板
	ControlTypeCollapsePanels byte = 7
	// ControlTypeGetClipboard 获取剪贴板
	ControlTypeGetClipboard byte = 8
	// ControlTypeSetClipboard 设置剪贴板
	ControlTypeSetClipboard byte = 9
	// ControlTypeSetDisplayPower 设置屏幕电源
	ControlTypeSetDisplayPower byte = 10
	// ControlTypeRotateDevice 旋转设备
	ControlTypeRotateDevice byte = 11
	// ControlTypeUHIDCreate 创建 UHID 设备
	ControlTypeUHIDCreate byte = 12
	// ControlTypeUHIDInput UHID 输入
	ControlTypeUHIDInput byte = 13
	// ControlTypeUHIDDestroy 销毁 UHID 设备
	ControlTypeUHIDDestroy byte = 14
	// ControlTypeOpenHardKeyboardSettings 打开硬件键盘设置
	ControlTypeOpenHardKeyboardSettings byte = 15
	// ControlTypeStartApp 启动应用
	ControlTypeStartApp byte = 16
	// ControlTypeResetVideo 重置视频
	ControlTypeResetVideo byte = 17
	// ControlTypeCameraSetTorch 设置相机闪光灯
	ControlTypeCameraSetTorch byte = 18
	// ControlTypeCameraZoomIn 相机放大
	ControlTypeCameraZoomIn byte = 19
	// ControlTypeCameraZoomOut 相机缩小
	ControlTypeCameraZoomOut byte = 20
	// ControlTypeResizeDisplay 调整显示器大小
	ControlTypeResizeDisplay byte = 21
	// ControlTypeScanFile 扫描文件
	ControlTypeScanFile byte = 22
)

// ws-scrcpy 扩展消息类型
const (
	// ControlTypeVideoSettings 视频设置动态调整
	ControlTypeVideoSettings byte = 101
	// ControlTypeFilePush 文件推送
	ControlTypeFilePush byte = 102
)

// 设备消息类型常量
const (
	// DeviceMsgClipboard 剪贴板内容变化
	DeviceMsgClipboard byte = 0
	// DeviceMsgAckClipboard 剪贴板确认
	DeviceMsgAckClipboard byte = 1
	// DeviceMsgUHIDOutput UHID 设备输出
	DeviceMsgUHIDOutput byte = 2
)

// 触摸事件动作常量
const (
	// TouchActionDown 按下
	TouchActionDown byte = 0
	// TouchActionUp 抬起
	TouchActionUp byte = 1
	// TouchActionMove 移动
	TouchActionMove byte = 2
	// TouchActionCancel 取消
	TouchActionCancel byte = 3
)

// 按键事件动作常量
const (
	// KeyActionDown 按下
	KeyActionDown byte = 0
	// KeyActionUp 抬起
	KeyActionUp byte = 1
	// KeyActionMulti 多键操作
	KeyActionMulti byte = 2
)

// 指针 ID 常量
const (
	// PointerIDMouse 鼠标指针
	PointerIDMouse uint64 = 0xFFFFFFFFFFFFFFFF
	// PointerIDGenericFinger 通用手指
	PointerIDGenericFinger uint64 = 0xFFFFFFFFFFFFFFFE
)

// 协议常量
const (
	// DeviceNameFieldLength 设备名字段长度
	DeviceNameFieldLength = 64
	// MaxControlMessageSize 控制消息最大大小
	MaxControlMessageSize = 1 << 18 // 256KB
	// MaxClipboardTextSize 剪贴板文本最大大小
	MaxClipboardTextSize = MaxControlMessageSize - 14
	// MaxTextLength 文本注入最大长度
	MaxTextLength = 300
	// MaxFileNameLength 文件名最大长度
	MaxFileNameLength = 256
	// MaxFilePathLength 文件路径最大长度
	MaxFilePathLength = 256
)
