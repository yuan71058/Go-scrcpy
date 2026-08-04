package scrcpy

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"

	"github.com/yuan71058/go-scrcpy/pkg/input"
	"github.com/yuan71058/go-scrcpy/pkg/render"
)

// SDLToAndroidKey SDL 键码到 Android 键码的映射
// 参考原版 scrcpy 的 input_manager.c: sc_input_manager_process_key
var SDLToAndroidKey = map[int32]int32{
	0x61: 29, 0x62: 30, 0x63: 31, 0x64: 32, 0x65: 33, // a-e
	0x66: 34, 0x67: 35, 0x68: 36, 0x69: 37, 0x6A: 38, // f-j
	0x6B: 39, 0x6C: 40, 0x6D: 41, 0x6E: 42, 0x6F: 43, // k-o
	0x70: 44, 0x71: 45, 0x72: 46, 0x73: 47, 0x74: 48, // p-t
	0x75: 49, 0x76: 50, 0x77: 51, 0x78: 52, 0x79: 53, // u-y
	0x7A: 54, // z
	0x30: 7, 0x31: 8, 0x32: 9, 0x33: 10, 0x34: 11, // 0-4
	0x35: 12, 0x36: 13, 0x37: 14, 0x38: 15, 0x39: 16, // 5-9
	0x4000004F: 3,  // HOME
	0x40000052: 4,  // BACK
	0x40000043: 24, // VOLUME_UP
	0x4000004E: 25, // VOLUME_DOWN
	0x40000044: 26, // POWER
	0x4000004B: 56, // MENU
	0x40000040: 66, // ENTER
	0x40000042: 67, // DELETE
	0x40000041: 61, // TAB
	0x4000004A: 111, // ESCAPE
	0x4000003D: 62, // SPACE
	0x40000051: 111, // ESCAPE (alt)
	0x400000E0: 21,  // LEFT
	0x400000E1: 22,  // RIGHT
	0x400000E2: 19,  // UP
	0x400000E3: 20,  // DOWN
}

// Display 投屏窗口
// 封装完整的 SDL 投屏流程：创建窗口、解码视频、处理输入
// 使用方式：
//
//	display, err := scrcpy.NewDisplay(serial, opts)
//	if err != nil { return }
//	defer display.Close()
//	display.Run() // 阻塞直到窗口关闭
type Display struct {
	client    *Client
	decoder   *render.H264Decoder
	displayW  int
	displayH  int
	serial    string
	closed    atomic.Bool

	// 帧同步（解码 goroutine → 主线程）
	frameData atomic.Value // []byte YUV 数据
	frameW    atomic.Int32
	frameH    atomic.Int32
	newFrame  atomic.Bool
}

// NewDisplay 创建投屏窗口
// 自动连接设备、创建解码器、初始化 SDL
func NewDisplay(serial string, opts Options) (*Display, error) {
	client := New(serial, opts)
	ctx := context.Background()

	fmt.Printf("连接设备 %s...\n", serial)
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("启动客户端失败: %w", err)
	}

	// 读取握手信息
	hs := client.Handshake()
	if hs == nil {
		client.Close()
		return nil, fmt.Errorf("读取握手数据失败")
	}

	displayW := int(hs.GetDisplayWidth())
	displayH := int(hs.GetDisplayHeight())
	fmt.Printf("视频尺寸: %dx%d\n", displayW, displayH)

	// 创建 FFmpeg 解码器
	decoder, err := render.NewH264Decoder()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("创建解码器失败: %w", err)
	}
	decoder.SetSize(displayW, displayH)

	// 初始化 SDL
	render.InitSDL()
	// 锁定 goroutine 到 OS 线程（SDL 窗口必须在其创建线程上处理消息）
	runtime.LockOSThread()

	d := &Display{
		client:   client,
		decoder:  decoder,
		displayW: displayW,
		displayH: displayH,
		serial:   serial,
	}

	// 启动解码 goroutine
	go d.decodeLoop()

	return d, nil
}

// decodeLoop 视频解码循环（goroutine）
func (d *Display) decodeLoop() {
	for frame := range d.client.VideoStream() {
		yuv, w, h, err := d.decoder.Decode(frame.Data)
		if err != nil || yuv == nil {
			continue
		}
		d.frameData.Store(yuv)
		d.frameW.Store(int32(w))
		d.frameH.Store(int32(h))
		d.newFrame.Store(true)
	}
}

// Run 运行投屏窗口（阻塞直到窗口关闭）
func (d *Display) Run() {
	var (
		sdl   *render.SDLRenderer
		lastW int32
		lastH int32
	)

	for {
		// 处理所有待处理事件
		for {
			evType, evData := render.PollEvent()
			if evType == 0 {
				break
			}
			switch evType {
			case render.SDL_EVENT_QUIT, render.SDL_EVENT_WINDOW_CLOSE_REQUESTED:
				fmt.Println("窗口关闭")
				return

			case render.SDL_EVENT_MOUSE_BUTTON_DOWN:
				btn := render.GetEventMouseButton(evData)
				x, y := render.GetEventMouseCoords(evData)
				d.client.SendControl(input.MouseDown(int32(x), int32(y), uint16(d.displayW), uint16(d.displayH), int32(btn)))

			case render.SDL_EVENT_MOUSE_BUTTON_UP:
				btn := render.GetEventMouseButton(evData)
				x, y := render.GetEventMouseCoords(evData)
				d.client.SendControl(input.MouseUp(int32(x), int32(y), uint16(d.displayW), uint16(d.displayH), int32(btn)))

			case render.SDL_EVENT_MOUSE_MOTION:
				x, y := render.GetEventMouseCoords(evData)
				d.client.SendControl(input.MouseMove(int32(x), int32(y), uint16(d.displayW), uint16(d.displayH), 0))

			case render.SDL_EVENT_MOUSE_WHEEL:
				hscroll, vscroll := render.GetEventMouseScroll(evData)
				x, y := render.GetEventMouseCoords(evData)
				d.client.SendControl(input.Scroll(int32(x), int32(y), uint16(d.displayW), uint16(d.displayH), int16(hscroll), int16(vscroll)))

			case render.SDL_EVENT_KEY_DOWN:
				keycode := render.GetEventKeycode(evData)
				mod := render.GetEventKeyMod(evData)
				if androidKey, ok := SDLToAndroidKey[keycode]; ok {
					d.client.SendControl(input.KeyDown(androidKey, mod))
				}

			case render.SDL_EVENT_KEY_UP:
				keycode := render.GetEventKeycode(evData)
				mod := render.GetEventKeyMod(evData)
				if androidKey, ok := SDLToAndroidKey[keycode]; ok {
					d.client.SendControl(input.KeyUp(androidKey, mod))
				}
			}
		}

		// 渲染新帧
		if d.newFrame.Load() {
			yuv := d.frameData.Load()
			w, h := d.frameW.Load(), d.frameH.Load()
			d.newFrame.Store(false)

			if yuv != nil {
				if w != lastW || h != lastH {
					if sdl != nil {
						sdl.Destroy()
					}
					title := fmt.Sprintf("scrcpy - %s (%dx%d)", d.serial, w, h)
					var err error
					sdl, err = render.NewSDLRenderer(title, int(w), int(h))
					if err != nil {
						fmt.Printf("创建渲染器失败: %v\n", err)
						return
					}
					lastW, lastH = w, h
					d.displayW, d.displayH = int(w), int(h)
				}
				if sdl != nil {
					if err := sdl.RenderYUV(yuv.([]byte)); err != nil {
						fmt.Printf("渲染失败: %v\n", err)
					}
				}
			}
		}

		// 让出 CPU，给 Windows 消息泵足够时间处理窗口消息
		render.Delay(16)
	}
}

// Close 关闭投屏窗口
func (d *Display) Close() {
	d.closed.Store(true)
	d.decoder.Close()
	d.client.Close()
}