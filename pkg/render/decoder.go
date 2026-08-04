package render

import (
	"fmt"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

var (
	avcodec *syscall.LazyDLL
	avutil  *syscall.LazyDLL

	procFindDecoderByID   *syscall.LazyProc
	procFindDecoderByName *syscall.LazyProc
	procAllocContext3     *syscall.LazyProc
	procOpen2             *syscall.LazyProc
	procSendPacket        *syscall.LazyProc
	procReceiveFrame      *syscall.LazyProc
	procFlushBuffers      *syscall.LazyProc
	procFreeContext       *syscall.LazyProc
	procPacketAlloc       *syscall.LazyProc
	procPacketFree        *syscall.LazyProc
	procPacketFromData    *syscall.LazyProc
	procPacketUnref       *syscall.LazyProc
	procFrameAlloc        *syscall.LazyProc
	procFrameFree         *syscall.LazyProc
	procMalloc            *syscall.LazyProc
	procAVFree            *syscall.LazyProc
)

const (
	AV_NOPTS_VALUE       int64 = -1 << 63
	AV_CODEC_FLAG_LOW_DELAY int32 = 0x800
)

func init() {
	dllDir := findFFmpegDir()
	if dllDir == "" {
		dllDir = "."
	}

	avcodec = loadDLL(filepath.Join(dllDir, "avcodec-62.dll"))
	avutil = loadDLL(filepath.Join(dllDir, "avutil-60.dll"))

	procFindDecoderByID = avcodec.NewProc("avcodec_find_decoder")
	procFindDecoderByName = avcodec.NewProc("avcodec_find_decoder_by_name")
	procAllocContext3 = avcodec.NewProc("avcodec_alloc_context3")
	procOpen2 = avcodec.NewProc("avcodec_open2")
	procSendPacket = avcodec.NewProc("avcodec_send_packet")
	procReceiveFrame = avcodec.NewProc("avcodec_receive_frame")
	procFlushBuffers = avcodec.NewProc("avcodec_flush_buffers")
	procFreeContext = avcodec.NewProc("avcodec_free_context")
	procPacketAlloc = avcodec.NewProc("av_packet_alloc")
	procPacketFree = avcodec.NewProc("av_packet_free")
	procPacketFromData = avcodec.NewProc("av_packet_from_data")
	procPacketUnref = avcodec.NewProc("av_packet_unref")
	procFrameAlloc = avutil.NewProc("av_frame_alloc")
	procFrameFree = avutil.NewProc("av_frame_free")
	procMalloc = avutil.NewProc("av_malloc")
	procAVFree = avutil.NewProc("av_free")
}

// H264Decoder 使用 FFmpeg avcodec 解码 H.264 视频流，输出 YUV420P 格式
type H264Decoder struct {
	codecCtx uintptr
	codec    uintptr
	frame    uintptr
	pkt      uintptr
	width    int
	height   int
	mu       sync.Mutex
	closed   bool
	opened   bool
}

// NewH264Decoder 创建 H264 解码器
func NewH264Decoder() (*H264Decoder, error) {
	d := &H264Decoder{}

	codec, _, _ := procFindDecoderByID.Call(uintptr(27)) // AV_CODEC_ID_H264
	if codec == 0 {
		name, _ := syscall.BytePtrFromString("h264")
		codec, _, _ = procFindDecoderByName.Call(uintptr(unsafe.Pointer(name)))
	}
	if codec == 0 {
		return nil, fmt.Errorf("找不到 H264 解码器")
	}
	d.codec = codec

	ctx, _, _ := procAllocContext3.Call(codec)
	if ctx == 0 {
		return nil, fmt.Errorf("分配上下文失败")
	}
	d.codecCtx = ctx

	f, _, _ := procFrameAlloc.Call()
	if f == 0 {
		procFreeContext.Call(uintptr(unsafe.Pointer(&d.codecCtx)))
		return nil, fmt.Errorf("分配帧失败")
	}
	d.frame = f

	p, _, _ := procPacketAlloc.Call()
	if p == 0 {
		procFreeContext.Call(uintptr(unsafe.Pointer(&d.codecCtx)))
		procFrameFree.Call(f)
		return nil, fmt.Errorf("分配 packet 失败")
	}
	d.pkt = p

	return d, nil
}

// open 打开解码器
func (d *H264Decoder) open() {
	if d.opened || d.codec == 0 {
		return
	}

	// 设置低延迟模式（参考原版 scrcpy）
	// flags 字段在 AVCodecContext 偏移 48 处
	flags := *(*int32)(unsafe.Pointer(d.codecCtx + 48))
	flags |= AV_CODEC_FLAG_LOW_DELAY
	*(*int32)(unsafe.Pointer(d.codecCtx + 48)) = flags

	if ret, _, _ := procOpen2.Call(d.codecCtx, d.codec, 0); ret == 0 {
		d.opened = true
	}
}

// Decode 解码 H264 数据，返回 YUV420P 格式数据
func (d *H264Decoder) Decode(data []byte) ([]byte, int, int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, 0, 0, fmt.Errorf("解码器已关闭")
	}
	if len(data) == 0 {
		return nil, 0, 0, nil
	}

	if !d.opened {
		d.open()
		if !d.opened {
			return nil, 0, 0, fmt.Errorf("打开解码器失败")
		}
	}

	buf, _, _ := procMalloc.Call(uintptr(len(data)))
	if buf == 0 {
		return nil, 0, 0, nil
	}
	for i := 0; i < len(data); i++ {
		*(*byte)(unsafe.Pointer(buf + uintptr(i))) = data[i]
	}

	procPacketUnref.Call(d.pkt)
	if ret, _, _ := procPacketFromData.Call(d.pkt, buf, uintptr(len(data))); ret != 0 {
		procAVFree.Call(buf)
		return nil, 0, 0, nil
	}

	*(*int64)(unsafe.Pointer(d.pkt + 8)) = AV_NOPTS_VALUE
	*(*int64)(unsafe.Pointer(d.pkt + 16)) = AV_NOPTS_VALUE

	sendRet, _, _ := procSendPacket.Call(d.codecCtx, d.pkt)
	if sendRet != 0 {
		// 发送失败时刷新解码器，防止错误状态累积
		procFlushBuffers.Call(d.codecCtx)
		return nil, 0, 0, nil
	}

	// 循环接收所有可用的帧（参考原版 scrcpy 的 decoder.c）
	for {
		recvRet, _, _ := procReceiveFrame.Call(d.codecCtx, d.frame)
		if recvRet == 0 {
			break // 成功收到一帧
		}
		code := int(int32(recvRet))
		if code == -11 { // AVERROR(EAGAIN) — 需要更多数据
			return nil, 0, 0, nil
		}
		// 其他错误，刷新解码器
		procFlushBuffers.Call(d.codecCtx)
		return nil, 0, 0, nil
	}

	// 使用已设置的宽高（从 handshake 获取）
	if d.width <= 0 || d.height <= 0 {
		return nil, 0, 0, nil
	}

	yPtr := *(*uintptr)(unsafe.Pointer(d.frame))
	uPtr := *(*uintptr)(unsafe.Pointer(d.frame + 8))
	vPtr := *(*uintptr)(unsafe.Pointer(d.frame + 16))
	yStride := int(*(*int32)(unsafe.Pointer(d.frame + 64)))
	uStride := int(*(*int32)(unsafe.Pointer(d.frame + 68)))
	vStride := int(*(*int32)(unsafe.Pointer(d.frame + 72)))

	ySize := d.width * d.height
	uvW := (d.width + 1) / 2
	uvH := (d.height + 1) / 2
	uvSize := uvW * uvH
	totalSize := ySize + uvSize*2

	yuv := make([]byte, totalSize)

	if yPtr != 0 {
		for i := 0; i < d.height; i++ {
			src := unsafe.Slice((*byte)(unsafe.Pointer(yPtr+uintptr(i*yStride))), d.width)
			copy(yuv[i*d.width:], src)
		}
	}
	if uPtr != 0 {
		for i := 0; i < uvH; i++ {
			src := unsafe.Slice((*byte)(unsafe.Pointer(uPtr+uintptr(i*uStride))), uvW)
			copy(yuv[ySize+i*uvW:], src)
		}
	}
	if vPtr != 0 {
		for i := 0; i < uvH; i++ {
			src := unsafe.Slice((*byte)(unsafe.Pointer(vPtr+uintptr(i*vStride))), uvW)
			copy(yuv[ySize+uvSize+i*uvW:], src)
		}
	}

	return yuv, d.width, d.height, nil
}

// SetSize 设置视频尺寸
func (d *H264Decoder) SetSize(w, h int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.width = w
	d.height = h
}

// GetSize 获取视频尺寸
func (d *H264Decoder) GetSize() (int, int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.width, d.height
}

// Close 关闭解码器
func (d *H264Decoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	if d.frame != 0 {
		procFrameFree.Call(uintptr(unsafe.Pointer(&d.frame)))
	}
	if d.pkt != 0 {
		procPacketFree.Call(uintptr(unsafe.Pointer(&d.pkt)))
	}
	if d.codecCtx != 0 {
		procFreeContext.Call(uintptr(unsafe.Pointer(&d.codecCtx)))
	}
	return nil
}