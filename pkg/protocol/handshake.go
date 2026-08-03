package protocol

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuan71058/go-scrcpy/pkg/transport"
	"github.com/yuan71058/go-scrcpy/pkg/types"
)

// Handshake 握手数据结构
// 包含设备名、显示器信息、编码器列表等
type Handshake struct {
	DeviceName    string          // 设备型号 (64 字节)
	Displays      []DisplayInfo   // 显示器信息列表
	Encoders      []string        // 可用编码器列表
	ClientID      int32           // 客户端 ID
	DeviceNameRaw [DeviceNameFieldLength]byte // 原始设备名数据
}

// DisplayInfo 显示器详细信息 (握手阶段解析)
type DisplayInfo struct {
	DisplayID       int32
	Width           int32
	Height          int32
	Rotation        int32
	LayerStack      int32
	Flags           int32
	ConnectionCount int32     // 当前连接数
	ScreenInfo      []byte    // ScreenInfo 数据 (25 字节)
	VideoSettings   []byte    // VideoSettings 数据 (35+ 字节)
}

// ReadHandshake 从流中读取并解析握手数据
// 视频流第一个消息包含设备元数据和流头
func ReadHandshake(reader *transport.ProtocolReader) (*Handshake, error) {
	logInfo("开始读取握手数据...")

	handshake := &Handshake{}

	// 读取设备名 (64 字节)
	deviceNameBytes, err := reader.ReadFull(DeviceNameFieldLength)
	if err != nil {
		return nil, fmt.Errorf("读取设备名失败: %w", err)
	}

	// 复制原始数据
	copy(handshake.DeviceNameRaw[:], deviceNameBytes)

	// 解析设备名 (UTF-8, null 填充)
	handshake.DeviceName = strings.TrimRight(string(deviceNameBytes), "\x00")
	logInfo("设备名: %s", handshake.DeviceName)

	// 读取视频流 codec ID (4 字节, big-endian)
	videoCodecID, err := reader.ReadUint32BE()
	if err != nil {
		return nil, fmt.Errorf("读取视频 codec ID 失败: %w", err)
	}
	logDebug("视频 codec ID: 0x%08X", videoCodecID)

	// 读取初始分辨率 (8 字节: width + height, big-endian)
	width, err := reader.ReadUint32BE()
	if err != nil {
		return nil, fmt.Errorf("读取宽度失败: %w", err)
	}
	height, err := reader.ReadUint32BE()
	if err != nil {
		return nil, fmt.Errorf("读取高度失败: %w", err)
	}

	logDebug("初始分辨率: %dx%d", width, height)

	logInfo("握手数据读取完成")
	return handshake, nil
}

// ReadHandshakeFromConn 从连接中读取握手数据
// 自动创建 ProtocolReader
func ReadHandshakeFromConn(conn io.Reader) (*Handshake, error) {
	reader := transport.NewProtocolReader(conn)
	return ReadHandshake(reader)
}

// GetDisplayInfo 根据显示器 ID 获取显示器信息
func (h *Handshake) GetDisplayInfo(displayID int32) *DisplayInfo {
	for i := range h.Displays {
		if h.Displays[i].DisplayID == displayID {
			return &h.Displays[i]
		}
	}
	return nil
}

// GetPrimaryDisplay 获取主显示器 (ID=0)
func (h *Handshake) GetPrimaryDisplay() *DisplayInfo {
	return h.GetDisplayInfo(0)
}

// GetVideoSize 获取指定显示器的视频尺寸
func (h *Handshake) GetVideoSize(displayID int32) (int32, int32) {
	display := h.GetDisplayInfo(displayID)
	if display == nil {
		return 0, 0
	}
	return display.Width, display.Height
}

// GetDeviceName 获取设备型号名称
func (h *Handshake) GetDeviceName() string {
	return h.DeviceName
}

// GetClientID 获取客户端 ID
func (h *Handshake) GetClientID() int32 {
	return h.ClientID
}

// String 返回握手信息的字符串表示
func (h *Handshake) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("设备名: %s\n", h.DeviceName))
	sb.WriteString(fmt.Sprintf("客户端 ID: %d\n", h.ClientID))
	sb.WriteString(fmt.Sprintf("显示器数量: %d\n", len(h.Displays)))

	for i, d := range h.Displays {
		sb.WriteString(fmt.Sprintf("  显示器 %d: ID=%d, %dx%d, rotation=%d, flags=0x%X\n",
			i, d.DisplayID, d.Width, d.Height, d.Rotation, d.Flags))
	}

	sb.WriteString(fmt.Sprintf("编码器数量: %d\n", len(h.Encoders)))
	for i, e := range h.Encoders {
		sb.WriteString(fmt.Sprintf("  编码器 %d: %s\n", i, e))
	}

	return sb.String()
}

// ReadDeviceNameOnly 仅读取设备名 (64 字节)
// 用于简单的设备识别场景
func ReadDeviceNameOnly(conn io.Reader) (string, error) {
	reader := transport.NewProtocolReader(conn)

	deviceNameBytes, err := reader.ReadFull(DeviceNameFieldLength)
	if err != nil {
		return "", fmt.Errorf("读取设备名失败: %w", err)
	}

	return strings.TrimRight(string(deviceNameBytes), "\x00"), nil
}

// ReadDisplayInfoFromBytes 从字节切片解析 DisplayInfo (24 字节)
func ReadDisplayInfoFromBytes(data []byte) (*types.DisplayInfo, error) {
	if len(data) < 24 {
		return nil, fmt.Errorf("DisplayInfo 数据不足: 需要 24 字节，实际 %d 字节", len(data))
	}

	// 大端序解析 6 个 int32
	info := &types.DisplayInfo{
		DisplayID:  int32(data[0])<<24 | int32(data[1])<<16 | int32(data[2])<<8 | int32(data[3]),
		Width:      int32(data[4])<<24 | int32(data[5])<<16 | int32(data[6])<<8 | int32(data[7]),
		Height:     int32(data[8])<<24 | int32(data[9])<<16 | int32(data[10])<<8 | int32(data[11]),
		Rotation:   int32(data[12])<<24 | int32(data[13])<<16 | int32(data[14])<<8 | int32(data[15]),
		LayerStack: int32(data[16])<<24 | int32(data[17])<<16 | int32(data[18])<<8 | int32(data[19]),
		Flags:      int32(data[20])<<24 | int32(data[21])<<16 | int32(data[22])<<8 | int32(data[23]),
	}

	return info, nil
}

// ReadScreenInfoFromBytes 从字节切片解析 ScreenInfo (25 字节)
func ReadScreenInfoFromBytes(data []byte) (*types.ScreenInfo, error) {
	if len(data) < 25 {
		return nil, fmt.Errorf("ScreenInfo 数据不足: 需要 25 字节，实际 %d 字节", len(data))
	}

	info := &types.ScreenInfo{
		ContentRect: types.Rect{
			Left:   int32(data[0])<<24 | int32(data[1])<<16 | int32(data[2])<<8 | int32(data[3]),
			Top:    int32(data[4])<<24 | int32(data[5])<<16 | int32(data[6])<<8 | int32(data[7]),
			Right:  int32(data[8])<<24 | int32(data[9])<<16 | int32(data[10])<<8 | int32(data[11]),
			Bottom: int32(data[12])<<24 | int32(data[13])<<16 | int32(data[14])<<8 | int32(data[15]),
		},
		VideoSize: types.Size{
			Width:  int32(data[16])<<24 | int32(data[17])<<16 | int32(data[18])<<8 | int32(data[19]),
			Height: int32(data[20])<<24 | int32(data[21])<<16 | int32(data[22])<<8 | int32(data[23]),
		},
		DeviceRotation: data[24],
	}

	return info, nil
}
