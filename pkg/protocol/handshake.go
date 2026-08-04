package protocol

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuan71058/go-scrcpy/pkg/transport"
)

// Handshake 握手数据结构
// 包含设备名、显示器信息、编码器列表等
type Handshake struct {
	DeviceName    string          // 设备型号 (64 字节)
	DisplayWidth  uint32          // 显示宽度
	DisplayHeight uint32          // 显示高度
	VideoCodecID  uint32          // 视频编解码器 ID (如 0x68323634 = "h264")
	DeviceNameRaw [DeviceNameFieldLength]byte // 原始设备名数据
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
	handshake.VideoCodecID = videoCodecID
	logDebug("视频 codec ID: 0x%08X", videoCodecID)

	// 读取 session packet (12 字节)
	// byte 0: 10000000 (session packet flag)
	// bytes 1-3: padding
	// bytes 4-7: video width (big-endian)
	// bytes 8-11: video height (big-endian)
	sessionPacket, err := reader.ReadFull(12)
	if err != nil {
		return nil, fmt.Errorf("读取 session packet 失败: %w", err)
	}

	// 解析宽度和高度 (bytes 4-7 和 8-11, big-endian)
	handshake.DisplayWidth = uint32(sessionPacket[4])<<24 | uint32(sessionPacket[5])<<16 |
		uint32(sessionPacket[6])<<8 | uint32(sessionPacket[7])
	handshake.DisplayHeight = uint32(sessionPacket[8])<<24 | uint32(sessionPacket[9])<<16 |
		uint32(sessionPacket[10])<<8 | uint32(sessionPacket[11])

	logDebug("初始分辨率: %dx%d", handshake.DisplayWidth, handshake.DisplayHeight)

	logInfo("握手数据读取完成")
	return handshake, nil
}

// ReadHandshakeFromConn 从连接中读取握手数据
// 自动创建 ProtocolReader
func ReadHandshakeFromConn(conn io.Reader) (*Handshake, error) {
	reader := transport.NewProtocolReader(conn)
	return ReadHandshake(reader)
}

// GetDeviceName 获取设备型号名称
func (h *Handshake) GetDeviceName() string {
	return h.DeviceName
}

// GetDisplayWidth 获取显示宽度
func (h *Handshake) GetDisplayWidth() uint32 {
	return h.DisplayWidth
}

// GetDisplayHeight 获取显示高度
func (h *Handshake) GetDisplayHeight() uint32 {
	return h.DisplayHeight
}

// GetVideoCodecID 获取视频编解码器 ID
func (h *Handshake) GetVideoCodecID() uint32 {
	return h.VideoCodecID
}

// GetVideoCodecName 获取视频编解码器名称
func (h *Handshake) GetVideoCodecName() string {
	switch h.VideoCodecID {
	case 0x68323634:
		return "h264"
	case 0x68323635:
		return "h265"
	case 0x00617631:
		return "av1"
	case 0x00767038:
		return "vp8"
	case 0x00767039:
		return "vp9"
	default:
		return fmt.Sprintf("unknown(0x%08X)", h.VideoCodecID)
	}
}

// String 返回握手信息的字符串表示
func (h *Handshake) String() string {
	return fmt.Sprintf("设备名: %s, 分辨率: %dx%d, 编码: %s", h.DeviceName, h.DisplayWidth, h.DisplayHeight, h.GetVideoCodecName())
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
