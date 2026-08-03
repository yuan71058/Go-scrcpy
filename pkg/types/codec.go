package types

// VideoCodec 视频编解码器类型
type VideoCodec int

const (
	// VideoCodecH264 H.264/AVC 编解码器 (默认)
	VideoCodecH264 VideoCodec = iota
	// VideoCodecH265 H.265/HEVC 编解码器
	VideoCodecH265
	// VideoCodecAV1 AV1 编解码器
	VideoCodecAV1
	// VideoCodecVP8 VP8 编解码器
	VideoCodecVP8
	// VideoCodecVP9 VP9 编解码器
	VideoCodecVP9
)

// String 返回编解码器的字符串表示
func (c VideoCodec) String() string {
	switch c {
	case VideoCodecH264:
		return "h264"
	case VideoCodecH265:
		return "h265"
	case VideoCodecAV1:
		return "av1"
	case VideoCodecVP8:
		return "vp8"
	case VideoCodecVP9:
		return "vp9"
	default:
		return "unknown"
	}
}

// AudioCodec 音频编解码器类型
type AudioCodec int

const (
	// AudioCodecOpus Opus 编解码器 (默认)
	AudioCodecOpus AudioCodec = iota
	// AudioCodecAAC AAC 编解码器
	AudioCodecAAC
	// AudioCodecFLAC FLAC 编解码器
	AudioCodecFLAC
	// AudioCodecRAW PCM 原始音频编解码器
	AudioCodecRAW
)

// String 返回编解码器的字符串表示
func (c AudioCodec) String() string {
	switch c {
	case AudioCodecOpus:
		return "opus"
	case AudioCodecAAC:
		return "aac"
	case AudioCodecFLAC:
		return "flac"
	case AudioCodecRAW:
		return "raw"
	default:
		return "unknown"
	}
}
