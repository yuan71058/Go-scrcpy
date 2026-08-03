package video

import (
	"bytes"
	"fmt"
)

// NALU 类型常量 (H.264)
const (
	NALUTypeH264Unspecified = 0
	NALUTypeH264NonIDR      = 1  // 非 IDR 帧
	NALUTypeH264PartitionA  = 2  // 分区 A
	NALUTypeH264PartitionB  = 3  // 分区 B
	NALUTypeH264PartitionC  = 4  // 分区 C
	NALUTypeH264IDR         = 5  // IDR 帧 (关键帧)
	NALUTypeH264SEI         = 6  // SEI 补充增强信息
	NALUTypeH264SPS         = 7  // SPS 序列参数集
	NALUTypeH264PPS         = 8  // PPS 图像参数集
	NALUTypeH264AUD         = 9  // AUD 访问单元分隔符
	NALUTypeH264Filler      = 12 // 填充数据
)

// NALU 类型常量 (H.265/HEVC)
const (
	NALUTypeH265VPS         = 32 // VPS 视频参数集
	NALUTypeH265SPS         = 33 // SPS 序列参数集
	NALUTypeH265PPS         = 34 // PPS 图像参数集
	NALUTypeH265IDR_W_RADL  = 19 // IDR 帧 (W_RADL)
	NALUTypeH265IDR_N_LP    = 20 // IDR 帧 (N_LP)
	NALUTypeH265Trail_R     = 1  // Trail R (非关键帧)
	NALUTypeH265Trail_N     = 0  // Trail N (非关键帧)
)

// NALU 表示一个 NAL 单元
type NALU struct {
	Type     NALUType // NALU 类型
	Data     []byte   // NALU 数据 (不含 start code)
	Temporal int      // 时域层 ID
}

// NALUType NALU 类型
type NALUType int

// H.264 NALU 类型名称映射
var h264NALUTypeNames = map[NALUType]string{
	NALUTypeH264Unspecified: "Unspecified",
	NALUTypeH264NonIDR:      "NonIDR",
	NALUTypeH264PartitionA:  "PartitionA",
	NALUTypeH264PartitionB:  "PartitionB",
	NALUTypeH264PartitionC:  "PartitionC",
	NALUTypeH264IDR:         "IDR",
	NALUTypeH264SEI:         "SEI",
	NALUTypeH264SPS:         "SPS",
	NALUTypeH264PPS:         "PPS",
	NALUTypeH264AUD:         "AUD",
	NALUTypeH264Filler:      "Filler",
}

// String 返回 NALU 类型名称
func (t NALUType) String() string {
	if name, ok := h264NALUTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", int(t))
}

// IsConfig 检查是否为配置 NALU (SPS/PPS/VPS)
func (t NALUType) IsConfig() bool {
	return t == NALUTypeH264SPS || t == NALUTypeH264PPS ||
		t == NALUTypeH265VPS || t == NALUTypeH265SPS || t == NALUTypeH265PPS
}

// IsKeyFrame 检查是否为关键帧 NALU
func (t NALUType) IsKeyFrame() bool {
	return t == NALUTypeH264IDR || t == NALUTypeH265IDR_W_RADL || t == NALUTypeH265IDR_N_LP
}

// NALUReader NALU 读取器
// 从 Annex B 格式的字节流中解析 NALU
type NALUReader struct {
	data   []byte // 原始数据
	offset int    // 当前偏移
}

// NewNALUReader 创建 NALU 读取器
func NewNALUReader(data []byte) *NALUReader {
	return &NALUReader{
		data:   data,
		offset: 0,
	}
}

// Next 读取下一个 NALU
// 返回 NALU 和是否还有更多数据
func (r *NALUReader) Next() (*NALU, bool) {
	for r.offset < len(r.data) {
		// 查找 start code (0x00000001 或 0x000001)
		startCodeLen := r.findStartCode()
		if startCodeLen == 0 {
			r.offset = len(r.data)
			return nil, false
		}

		// 记录 NALU 数据起始位置
		naluStart := r.offset

		// 查找下一个 start code
		nextStart := r.findNextStartCode()
		naluEnd := nextStart
		if nextStart == 0 {
			naluEnd = len(r.data)
		}

		// 提取 NALU 数据（含 start code）
		naluData := r.data[naluStart:naluEnd]
		r.offset = naluEnd

		// 解析 NALU
		nalu := r.parseNALU(naluData, startCodeLen)
		if nalu != nil {
			return nalu, true
		}
	}

	return nil, false
}

// findStartCode 查找 start code 长度
// 返回 3 (0x000001) 或 4 (0x00000001)，0 表示未找到
func (r *NALUReader) findStartCode() int {
	for i := r.offset; i < len(r.data)-2; i++ {
		if r.data[i] == 0 && r.data[i+1] == 0 {
			if i+2 < len(r.data) && r.data[i+2] == 1 {
				return 3
			}
			if i+3 < len(r.data) && r.data[i+2] == 0 && r.data[i+3] == 1 {
				return 4
			}
		}
	}
	return 0
}

// findNextStartCode 查找下一个 start code 的位置
func (r *NALUReader) findNextStartCode() int {
	start := r.offset + 3
	if start >= len(r.data) {
		return 0
	}

	for i := start; i < len(r.data)-2; i++ {
		if r.data[i] == 0 && r.data[i+1] == 0 {
			if i+2 < len(r.data) && r.data[i+2] == 1 {
				return i
			}
			if i+3 < len(r.data) && r.data[i+2] == 0 && r.data[i+3] == 1 {
				return i
			}
		}
	}
	return 0
}

// parseNALU 解析 NALU 数据
func (r *NALUReader) parseNALU(data []byte, startCodeLen int) *NALU {
	if len(data) <= startCodeLen {
		return nil
	}

	// 跳过 start code，读取 NALU header
	headerByte := data[startCodeLen]

	// H.264: 低 5 位为类型
	// H.265: 高 1 位为 0，后面 6 位为类型，再低 1 位为层 ID
	naluType := NALUType(headerByte & 0x1F)

	return &NALU{
		Type:     naluType,
		Data:     data[startCodeLen:],
		Temporal: int((headerByte >> 5) & 0x03),
	}
}

// ParseSPSInfo 解析 H.264 SPS 信息
// 提取宽高等基本信息
type SPSInfo struct {
	Width       int
	Height      int
	ProfileIDC  int
	LevelIDC    int
	ChromaFmtIDC int
}

// ParseSPS 解析 H.264 SPS NALU
// 简化实现，仅提取宽高
func ParseSPS(data []byte) (*SPSInfo, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("SPS 数据过短: %d bytes", len(data))
	}

	info := &SPSInfo{}

	// 跳过 NALU header (1 字节)
	offset := 1

	// profile_idc (8 bits)
	info.ProfileIDC = int(data[offset])
	offset++

	// constraint_set0-5_flag + reserved_zero_2bits (8 bits)
	offset++

	// level_idc (8 bits)
	info.LevelIDC = int(data[offset])
	offset++

	// seq_parameter_set_id (ue(v))
	// 跳过后续复杂解析，使用简化方式
	// 实际应用中应使用完整的 Exp-Golomb 解码器

	// 返回基本信息
	info.Width = 1920  // 默认值，实际应从 SPS 解析
	info.Height = 1080 // 默认值

	return info, nil
}

// SplitFrameData 将编码帧数据分割为独立的 NALU 列表
// 输入: Annex B 格式的帧数据
// 输出: NALU 列表（不含 start code）
func SplitFrameData(data [][]byte) [][]byte {
	var nalus [][]byte

	for _, frame := range data {
		reader := NewNALUReader(frame)
		for {
			nalu, ok := reader.Next()
			if !ok {
				break
			}
			nalus = append(nalus, nalu.Data)
		}
	}

	return nalus
}

// HasSPSPPS 检查帧数据是否包含 SPS/PPS
func HasSPSPPS(data []byte) bool {
	reader := NewNALUReader(data)
	for {
		nalu, ok := reader.Next()
		if !ok {
			break
		}
		if nalu.Type == NALUTypeH264SPS || nalu.Type == NALUTypeH264PPS ||
			nalu.Type == NALUTypeH265VPS || nalu.Type == NALUTypeH265SPS || nalu.Type == NALUTypeH265PPS {
			return true
		}
	}
	return false
}

// IsKeyFrame 检查帧数据是否为关键帧
func IsKeyFrame(data []byte) bool {
	reader := NewNALUReader(data)
	for {
		nalu, ok := reader.Next()
		if !ok {
			break
		}
		if nalu.Type.IsKeyFrame() {
			return true
		}
	}
	return false
}

// ConcatNALUs 连接多个 NALU 为 Annex B 格式
func ConcatNALUs(nalus [][]byte) []byte {
	var buf bytes.Buffer
	for _, nalu := range nalus {
		buf.Write([]byte{0x00, 0x00, 0x00, 0x01}) // start code
		buf.Write(nalu)
	}
	return buf.Bytes()
}
