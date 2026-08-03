package protocol

import (
	"fmt"

	"github.com/yuan71058/go-scrcpy/pkg/transport"
)

// FrameHeader 帧元数据头 (12 字节)
// 每个编码包前附带，包含 PTS 时间戳和包大小
type FrameHeader struct {
	PTS    int64  // 显示时间戳 (微秒)
	Size   uint32 // 编码包大小 (字节)
	Config bool   // 是否为配置帧 (SPS/PPS)
	Key    bool   // 是否为关键帧
}

// ReadFrameHeader 从流中读取帧元数据头
// 协议格式 (12 字节):
// [8B] PTS + flags (big-endian):
//   bit 63 = PACKET_FLAG_CONFIG (配置包)
//   bit 62 = PACKET_FLAG_KEY_FRAME (关键帧)
//   其余位 = PTS 微秒时间戳
// [4B] Packet size (big-endian)
func ReadFrameHeader(reader *transport.ProtocolReader) (*FrameHeader, error) {
	logDebug("读取帧元数据头...")

	// 读取 PTS + flags (big-endian)
	ptsAndFlags, err := reader.ReadUint64BE()
	if err != nil {
		return nil, fmt.Errorf("读取 PTS+flags 失败: %w", err)
	}

	// 读取包大小 (big-endian)
	size, err := reader.ReadUint32BE()
	if err != nil {
		return nil, fmt.Errorf("读取包大小失败: %w", err)
	}

	// 解析标志位
	config := (ptsAndFlags & PacketFlagConfig) != 0
	key := (ptsAndFlags & PacketFlagKeyFrame) != 0

	// 提取 PTS (使用掩码清除标志位)
	pts := ptsAndFlags & PacketPTSMask

	header := &FrameHeader{
		PTS:    int64(pts),
		Size:   size,
		Config: config,
		Key:    key,
	}

	logDebug("帧头: PTS=%d, size=%d, config=%v, key=%v", pts, size, config, key)

	return header, nil
}

// IsConfigPacket 检查是否为配置包 (SPS/PPS)
func (h *FrameHeader) IsConfigPacket() bool {
	return h.Config
}

// IsKeyFrame 检查是否为关键帧
func (h *FrameHeader) IsKeyFrame() bool {
	return h.Key
}

// GetPTS 获取 PTS 时间戳 (微秒)
func (h *FrameHeader) GetPTS() int64 {
	return h.PTS
}

// GetSize 获取编码包大小 (字节)
func (h *FrameHeader) GetSize() uint32 {
	return h.Size
}

// String 返回帧头的字符串表示
func (h *FrameHeader) String() string {
	flags := ""
	if h.Config {
		flags += " CONFIG"
	}
	if h.Key {
		flags += " KEY"
	}
	return fmt.Sprintf("FrameHeader(PTS=%d, size=%d%s)", h.PTS, h.Size, flags)
}

// FrameHeaderSize 帧元数据头大小 (字节)
const FrameHeaderSize = 12

// ReadFrameData 读取帧数据
// 先读取帧头，再读取对应的编码数据
func ReadFrameData(reader *transport.ProtocolReader) (*FrameHeader, []byte, error) {
	// 读取帧头
	header, err := ReadFrameHeader(reader)
	if err != nil {
		return nil, nil, err
	}

	// 读取编码数据
	data, err := reader.ReadFull(int(header.Size))
	if err != nil {
		return nil, nil, fmt.Errorf("读取帧数据失败: %w", err)
	}

	return header, data, nil
}

// SkipFrameData 跳过帧数据
// 读取帧头后跳过对应的编码数据
func SkipFrameData(reader *transport.ProtocolReader) (*FrameHeader, error) {
	header, err := ReadFrameHeader(reader)
	if err != nil {
		return nil, err
	}

	// 跳过编码数据
	_, err = reader.ReadFull(int(header.Size))
	if err != nil {
		return nil, fmt.Errorf("跳过帧数据失败: %w", err)
	}

	return header, nil
}
