// 投屏示例
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/input"
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
	"github.com/yuan71058/go-scrcpy/pkg/render"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if exe, err := os.Executable(); err == nil {
		dir = filepath.Dir(exe)
		for i := 0; i < 10; i++ {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

var sdlToAndroidKey = map[int32]int32{
	0x61: 29, 0x62: 30, 0x63: 31, 0x64: 32, 0x65: 33,
	0x66: 34, 0x67: 35, 0x68: 36, 0x69: 37, 0x6A: 38,
	0x6B: 39, 0x6C: 40, 0x6D: 41, 0x6E: 42, 0x6F: 43,
	0x70: 44, 0x71: 45, 0x72: 46, 0x73: 47, 0x74: 48,
	0x75: 49, 0x76: 50, 0x77: 51, 0x78: 52, 0x79: 53,
	0x7A: 54,
	0x30: 7, 0x31: 8, 0x32: 9, 0x33: 10, 0x34: 11,
	0x35: 12, 0x36: 13, 0x37: 14, 0x38: 15, 0x39: 16,
	0x4000004F: 3, 0x40000052: 4, 0x40000043: 24, 0x4000004E: 25,
	0x40000044: 26, 0x4000004B: 56, 0x40000040: 66, 0x40000042: 67,
	0x40000041: 61, 0x4000004A: 111, 0x4000003D: 62, 0x40000051: 111,
	0x400000E0: 21, 0x400000E1: 22, 0x400000E2: 19, 0x400000E3: 20,
}

func main() {
	adbClient := adb.NewClient("adb")
	ctx := context.Background()

	devices, err := adbClient.ListDevices(ctx)
	if err != nil {
		log.Fatalf("列举设备失败: %v", err)
	}
	if len(devices) == 0 {
		log.Fatal("未找到设备")
	}
	serial := devices[0].Serial
	fmt.Printf("设备: %s\n", serial)

	rootCmd := exec.Command("adb", "-s", serial, "root")
	rootCmd.CombinedOutput()

	projectRoot := findProjectRoot()
	if projectRoot == "" {
		log.Fatal("无法找到项目根目录 (go.mod)")
	}
	localJAR := filepath.Join(projectRoot, "data", "scrcpy-server.jar")
	fmt.Printf("JAR: %s\n", localJAR)

	opts := scrcpy.DefaultOptions()
	opts.LocalJAR = localJAR
	opts.Server.VideoBitRate = 4000000
	opts.Server.Audio = false
	opts.Server.Control = true
	opts.Server.VideoCodec = "h264"

	listenPort := 27183
	client := scrcpy.New(serial, opts, listenPort)

	fmt.Println("连接设备...")
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
	defer client.Close()

	decoder, err := render.NewH264Decoder()
	if err != nil {
		log.Fatalf("创建解码器失败: %v", err)
	}
	defer decoder.Close()

	var displayW, displayH int
	if hs := client.Handshake(); hs != nil {
		if h, ok := hs.(*protocol.Handshake); ok {
			displayW = int(h.GetDisplayWidth())
			displayH = int(h.GetDisplayHeight())
			decoder.SetSize(displayW, displayH)
			codecName := h.GetVideoCodecName()
			fmt.Printf("视频尺寸: %dx%d\n", displayW, displayH)
			fmt.Printf("视频编码: %s (0x%08X)\n", codecName, h.GetVideoCodecID())
		}
	}

	fmt.Println("连接成功！启动投屏窗口...")

	render.InitSDL()
	// 锁定当前 goroutine 到 OS 线程，确保 SDL 窗口创建和事件处理在同一线程
	// 参考原版 scrcpy：SDL 窗口必须在其创建线程上处理消息
	runtime.LockOSThread()

	// 用 atomic 同步帧数据，避免锁
	var (
		frameData atomic.Value // []byte
		frameW    atomic.Int32
		frameH    atomic.Int32
		newFrame  atomic.Bool
	)

	// 视频解码 goroutine
	go func() {
		for frame := range client.VideoStream() {
			nv12, w, h, err := decoder.Decode(frame.Data)
			if err != nil || nv12 == nil {
				continue
			}
			frameData.Store(nv12)
			frameW.Store(int32(w))
			frameH.Store(int32(h))
			newFrame.Store(true)
		}
	}()

	// 主线程事件循环
	// 使用 PollEvent 轮询 + Delay(16) 避免忙等导致 Windows 消息泵饥饿
	// 原版 scrcpy 使用 SDL_WaitEvent 阻塞等待 + PushEvent 唤醒，但 Go 窗口在首帧前已显示
	// 若用 WaitEvent 会导致首帧到达前窗口无响应，因此改用 PollEvent + 适当延迟
	var (
		sdl        *render.SDLRenderer
		lastW      int32
		lastH      int32
		frameCount int
	)
	for {
		// 1. 处理所有待处理事件
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
				client.SendControl(input.MouseDown(int32(x), int32(y), uint16(displayW), uint16(displayH), int32(btn)))
			case render.SDL_EVENT_MOUSE_BUTTON_UP:
				btn := render.GetEventMouseButton(evData)
				x, y := render.GetEventMouseCoords(evData)
				client.SendControl(input.MouseUp(int32(x), int32(y), uint16(displayW), uint16(displayH), int32(btn)))
			case render.SDL_EVENT_MOUSE_MOTION:
				x, y := render.GetEventMouseCoords(evData)
				client.SendControl(input.MouseMove(int32(x), int32(y), uint16(displayW), uint16(displayH), 0))
			case render.SDL_EVENT_MOUSE_WHEEL:
				hscroll, vscroll := render.GetEventMouseScroll(evData)
				x, y := render.GetEventMouseCoords(evData)
				client.SendControl(input.Scroll(int32(x), int32(y), uint16(displayW), uint16(displayH), int16(hscroll), int16(vscroll)))
			case render.SDL_EVENT_KEY_DOWN:
				keycode := render.GetEventKeycode(evData)
				mod := render.GetEventKeyMod(evData)
				if androidKey, ok := sdlToAndroidKey[keycode]; ok {
					client.SendControl(input.KeyDown(androidKey, mod))
				}
			case render.SDL_EVENT_KEY_UP:
				keycode := render.GetEventKeycode(evData)
				mod := render.GetEventKeyMod(evData)
				if androidKey, ok := sdlToAndroidKey[keycode]; ok {
					client.SendControl(input.KeyUp(androidKey, mod))
				}
			}
		}

		// 2. 渲染新帧
		if newFrame.Load() {
			nv12 := frameData.Load()
			w, h := frameW.Load(), frameH.Load()
			newFrame.Store(false)

			if nv12 != nil {
				if w != lastW || h != lastH {
					if sdl != nil {
						sdl.Destroy()
					}
					title := fmt.Sprintf("scrcpy - %s (%dx%d)", serial, w, h)
					var err error
					sdl, err = render.NewSDLRenderer(title, int(w), int(h))
					if err != nil {
						log.Fatalf("创建渲染器失败: %v", err)
					}
					lastW, lastH = w, h
					displayW, displayH = int(w), int(h)
					fmt.Printf("视频尺寸: %dx%d\n", w, h)
				}
				if sdl != nil {
					if err := sdl.RenderYUV(nv12.([]byte)); err != nil {
						log.Printf("渲染失败: %v", err)
					}
				}
				frameCount++
				if frameCount%100 == 0 {
					fmt.Printf("已渲染 %d 帧\n", frameCount)
				}
			}
		}

		// 3. 让出 CPU，给 Windows 消息泵足够时间处理窗口消息
		// 16ms ≈ 60fps，避免原 1ms 延迟导致消息泵饥饿
		render.Delay(16)
	}
}