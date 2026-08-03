package audio

import (
	"fmt"
	"sync"
)

// PCMPlayer PCM 音频播放器
// 基础实现：将 PCM 数据写入内部缓冲区
// 生产环境可替换为 SDL/PulseAudio 绑定
type PCMPlayer struct {
	buffer chan []byte // PCM 数据缓冲区
	mu     sync.Mutex
	closed bool
}

// NewPCMPlayer 创建 PCM 播放器
// capacity: 缓冲区容量
func NewPCMPlayer(capacity int) *PCMPlayer {
	if capacity <= 0 {
		capacity = 100
	}
	return &PCMPlayer{
		buffer: make(chan []byte, capacity),
	}
}

// Play 播放 PCM 数据
func (p *PCMPlayer) Play(pcm []byte) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("播放器已关闭")
	}
	p.mu.Unlock()

	// 复制数据
	data := make([]byte, len(pcm))
	copy(data, pcm)

	// 放入缓冲区
	select {
	case p.buffer <- data:
		logDebug("播放 PCM: size=%d", len(pcm))
		return nil
	default:
		// 缓冲区满，丢弃
		logDebug("PCM 缓冲区满，丢弃数据")
		return nil
	}
}

// ReadBuffer 读取缓冲区中的 PCM 数据（非阻塞）
func (p *PCMPlayer) ReadBuffer() ([]byte, bool) {
	select {
	case data := <-p.buffer:
		return data, true
	default:
		return nil, false
	}
}

// Close 关闭播放器
func (p *PCMPlayer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.buffer)
	logInfo("PCM 播放器已关闭")
	return nil
}
