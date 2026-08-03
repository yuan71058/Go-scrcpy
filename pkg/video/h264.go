package video

import (
	"fmt"
	"sync"
)

// H264Decoder H.264 解码器
// 基础实现：存储编码帧数据
// 生产环境可替换为 FFmpeg CGo 绑定或 joy5
type H264Decoder struct {
	frames  chan *DecodedFrame // 解码后的帧队列
	width   int               // 视频宽度
	height  int               // 视频高度
	mu      sync.Mutex
	closed  bool
	spsData []byte // SPS 数据
	ppsData []byte // PPS 数据
}

// NewH264Decoder 创建 H.264 解码器
// capacity: 帧队列容量
func NewH264Decoder(capacity int) *H264Decoder {
	if capacity <= 0 {
		capacity = 60
	}
	return &H264Decoder{
		frames: make(chan *DecodedFrame, capacity),
	}
}

// Push 推送编码数据到解码器
// 数据应为 Annex B 格式（含 start code）
func (d *H264Decoder) Push(data []byte) error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return fmt.Errorf("解码器已关闭")
	}
	d.mu.Unlock()

	// 检查是否为配置帧 (SPS/PPS)
	isConfig := HasSPSPPS(data)
	isKey := IsKeyFrame(data)

	// 解析 NALU 获取类型
	reader := NewNALUReader(data)
	for {
		nalu, ok := reader.Next()
		if !ok {
			break
		}

		// 保存 SPS/PPS
		if nalu.Type == NALUTypeH264SPS {
			d.spsData = make([]byte, len(nalu.Data))
			copy(d.spsData, nalu.Data)
		} else if nalu.Type == NALUTypeH264PPS {
			d.ppsData = make([]byte, len(nalu.Data))
			copy(d.ppsData, nalu.Data)
		}
	}

	// 创建帧对象
	frame := &DecodedFrame{
		PTS:    0, // 由调用者设置
		Data:   make([]byte, len(data)),
		Width:  d.width,
		Height: d.height,
		Config: isConfig,
		Key:    isKey,
	}
	copy(frame.Data, data)

	// 尝试放入队列
	select {
	case d.frames <- frame:
		logDebug("推送帧: config=%v, key=%v, size=%d", isConfig, isKey, len(data))
		return nil
	default:
		// 队列满，丢弃最旧的帧
		select {
		case <-d.frames:
			logDebug("帧队列满，丢弃旧帧")
		default:
		}
		d.frames <- frame
		return nil
	}
}

// ReadFrame 读取解码后的帧（阻塞）
func (d *H264Decoder) ReadFrame() (*DecodedFrame, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil, fmt.Errorf("解码器已关闭")
	}
	d.mu.Unlock()

	frame := <-d.frames
	if frame == nil {
		return nil, fmt.Errorf("解码器已关闭")
	}
	return frame, nil
}

// Close 关闭解码器
func (d *H264Decoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	close(d.frames)
	logInfo("H.264 解码器已关闭")
	return nil
}

// SetSize 设置视频尺寸
func (d *H264Decoder) SetSize(width, height int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.width = width
	d.height = height
	logDebug("设置视频尺寸: %dx%d", width, height)
}

// GetSize 获取视频尺寸
func (d *H264Decoder) GetSize() (int, int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.width, d.height
}

// GetSPS 获取 SPS 数据
func (d *H264Decoder) GetSPS() []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.spsData == nil {
		return nil
	}
	result := make([]byte, len(d.spsData))
	copy(result, d.spsData)
	return result
}

// GetPPS 获取 PPS 数据
func (d *H264Decoder) GetPPS() []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ppsData == nil {
		return nil
	}
	result := make([]byte, len(d.ppsData))
	copy(result, d.ppsData)
	return result
}
