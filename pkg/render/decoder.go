// Package render 提供 FFmpeg H264 解码和 SDL3 渲染功能
package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

var (
	avcodec *syscall.LazyDLL
	avutil  *syscall.LazyDLL

	procFindDecoder       *syscall.LazyProc
	procAllocContext3     *syscall.LazyProc
	procOpen2             *syscall.LazyProc
	procSendPacket        *syscall.LazyProc
	procReceiveFrame      *syscall.LazyProc
	procFreeContext       *syscall.LazyProc
	procPacketAlloc       *syscall.LazyProc
	procPacketFree        *syscall.LazyProc
	procPacketFromData    *syscall.LazyProc
	procFrameAlloc        *syscall.LazyProc
	procFrameFree         *syscall.LazyProc
)

const (
	AV_CODEC_ID_H264 = 27
)

func init() {
	exe, _ := os.Executable()
	// mirror.exe 在 _examples/mirror/，向上三级到 project root
	root := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	dllDir := filepath.Join(root, "data", "scrcpy-win64-v4.1") + `\`

	avcodec = syscall.NewLazyDLL(dllDir + "avcodec-62.dll")
	avutil = syscall.NewLazyDLL(dllDir + "avutil-60.dll")

	procFindDecoder = avcodec.NewProc("avcodec_find_decoder")
	procAllocContext3 = avcodec.NewProc("avcodec_alloc_context3")
	procOpen2 = avcodec.NewProc("avcodec_open2")
	procSendPacket = avcodec.NewProc("avcodec_send_packet")
	procReceiveFrame = avcodec.NewProc("avcodec_receive_frame")
	procFreeContext = avcodec.NewProc("avcodec_free_context")
	procPacketAlloc = avcodec.NewProc("av_packet_alloc")
	procPacketFree = avcodec.NewProc("av_packet_free")
	procPacketFromData = avcodec.NewProc("av_packet_from_data")
	procFrameAlloc = avutil.NewProc("av_frame_alloc")
	procFrameFree = avutil.NewProc("av_frame_free")
}

// H264Decoder H264 解码器
type H264Decoder struct {
	codecCtx uintptr
	frame    uintptr
	pkt      uintptr
	width    int
	height   int
	mu       sync.Mutex
	closed   bool
}

// NewH264Decoder 创建 H264 解码器
func NewH264Decoder() (*H264Decoder, error) {
	d := &H264Decoder{}

	codec, _, _ := procFindDecoder.Call(uintptr(AV_CODEC_ID_H264))
	if codec == 0 {
		return nil, fmt.Errorf("找不到 H264 解码器")
	}

	ctx, _, _ := procAllocContext3.Call(codec)
	if ctx == 0 {
		return nil, fmt.Errorf("分配上下文失败")
	}
	d.codecCtx = ctx

	ret, _, _ := procOpen2.Call(ctx, codec, 0)
	if ret != 0 {
		procFreeContext.Call(uintptr(unsafe.Pointer(&d.codecCtx)))
		return nil, fmt.Errorf("打开解码器失败: %d", ret)
	}

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

// Decode 解码 H264 数据，返回 YUV420P 数据
func (d *H264Decoder) Decode(data []byte) ([]byte, int, int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, 0, 0, fmt.Errorf("解码器已关闭")
	}

	cData := make([]byte, len(data))
	copy(cData, data)
	procPacketFromData.Call(d.pkt, uintptr(unsafe.Pointer(&cData[0])), uintptr(len(data)))

	ret, _, _ := procSendPacket.Call(d.codecCtx, d.pkt)
	if ret != 0 {
		return nil, 0, 0, fmt.Errorf("发送数据失败: %d", ret)
	}

	ret, _, _ = procReceiveFrame.Call(d.codecCtx, d.frame)
	if ret != 0 {
		return nil, 0, 0, fmt.Errorf("接收帧失败: %d", ret)
	}

	w := int(*(*int32)(unsafe.Pointer(d.codecCtx + 48)))
	h := int(*(*int32)(unsafe.Pointer(d.codecCtx + 52)))
	if w <= 0 || h <= 0 {
		return nil, 0, 0, fmt.Errorf("无效尺寸: %dx%d", w, h)
	}
	d.width = w
	d.height = h

	// 复制 YUV420P 数据
	yPtr := *(*uintptr)(unsafe.Pointer(d.frame))
	uPtr := *(*uintptr)(unsafe.Pointer(d.frame + 8))
	vPtr := *(*uintptr)(unsafe.Pointer(d.frame + 16))
	yStride := int(*(*int32)(unsafe.Pointer(d.frame + 64)))
	uStride := int(*(*int32)(unsafe.Pointer(d.frame + 68)))
	vStride := int(*(*int32)(unsafe.Pointer(d.frame + 72)))

	ySize := w * h
	uvW := (w + 1) / 2
	uvH := (h + 1) / 2
	uvSize := uvW * uvH
	totalSize := ySize + uvSize*2

	yuv := make([]byte, totalSize)

	if yPtr != 0 {
		for i := 0; i < h; i++ {
			src := unsafe.Slice((*byte)(unsafe.Pointer(yPtr+uintptr(i*yStride))), w)
			copy(yuv[i*w:], src)
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

	return yuv, w, h, nil
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
		d.frame = 0
	}
	if d.pkt != 0 {
		procPacketFree.Call(uintptr(unsafe.Pointer(&d.pkt)))
		d.pkt = 0
	}
	if d.codecCtx != 0 {
		procFreeContext.Call(uintptr(unsafe.Pointer(&d.codecCtx)))
		d.codecCtx = 0
	}
	return nil
}
