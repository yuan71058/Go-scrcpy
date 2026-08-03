package audio

import (
	"fmt"
	"sync"
)

// OpusDecoder Opus 音频解码器
// 基础实现：存储编码帧数据
// 生产环境可替换为 opus 绑定实现
type OpusDecoder struct {
	packets chan *AudioData // 解码后的音频队列
	mu      sync.Mutex
	closed  bool
}

// NewOpusDecoder 创建 Opus 解码器
// capacity: 音频包队列容量
func NewOpusDecoder(capacity int) *OpusDecoder {
	if capacity <= 0 {
		capacity = 100
	}
	return &OpusDecoder{
		packets: make(chan *AudioData, capacity),
	}
}

// Push 推送编码数据到解码器
func (d *OpusDecoder) Push(data []byte) error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return fmt.Errorf("解码器已关闭")
	}
	d.mu.Unlock()

	// 创建音频数据对象
	pkt := &AudioData{
		PTS:  0, // 由调用者设置
		Data: make([]byte, len(data)),
	}
	copy(pkt.Data, data)

	// 尝试放入队列
	select {
	case d.packets <- pkt:
		logDebug("推送音频包: size=%d", len(data))
		return nil
	default:
		// 队列满，丢弃最旧的包
		select {
		case <-d.packets:
			logDebug("音频队列满，丢弃旧包")
		default:
		}
		d.packets <- pkt
		return nil
	}
}

// ReadPCM 读取解码后的 PCM 数据（阻塞）
func (d *OpusDecoder) ReadPCM() (*AudioData, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil, fmt.Errorf("解码器已关闭")
	}
	d.mu.Unlock()

	pkt := <-d.packets
	if pkt == nil {
		return nil, fmt.Errorf("解码器已关闭")
	}
	return pkt, nil
}

// Close 关闭解码器
func (d *OpusDecoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	close(d.packets)
	logInfo("Opus 解码器已关闭")
	return nil
}
