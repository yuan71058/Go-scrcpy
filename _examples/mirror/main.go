// 投屏示例 - 使用项目自己的 Go 代码实现投屏+控制
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

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
	opts.Server.VideoCodec = "h264" // 强制 H264 编码，FFmpeg 解码器只支持 H264

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

	// 必须先在主线程初始化 SDL，否则事件循环会失败
	render.InitSDL()

	var (
		mu          sync.Mutex
		sdl         *render.SDLRenderer
		frameData   []byte
		frameW      int
		frameH      int
		lastW, lastH int
	)

	// 视频解码 goroutine
	go func() {
		for frame := range client.VideoStream() {
			nv12, w, h, err := decoder.Decode(frame.Data)
			if err != nil || nv12 == nil {
				continue
			}

			mu.Lock()
			frameData = nv12
			frameW = w
			frameH = h
			mu.Unlock()
		// 通知主线程有新帧（原版 scrcpy 方式）
		render.PushEvent(render.SC_EVENT_NEW_FRAME, 0)
		}
	}()

	// 主线程事件循环：使用 WaitEvent（原版 scrcpy 方式，修复 PushEvent 后可用）
	frameCount := 0
	for {
		evType, evData := render.WaitEvent()
		switch evType {
		case 0:
			// WaitEvent 失败，短暂延迟后重试
			render.Delay(1)
			continue

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

		// 每收到一帧事件就渲染最新帧
		mu.Lock()
		nv12, w, h := frameData, frameW, frameH
		frameData = nil
		mu.Unlock()

		if nv12 != nil {
			if sdl == nil || w != lastW || h != lastH {
				if sdl != nil {
					sdl.Destroy()
				}
				title := fmt.Sprintf("scrcpy - %s (%dx%d)", serial, w, h)
				sdl, err = render.NewSDLRenderer(title, w, h)
				if err != nil {
					log.Fatalf("创建渲染器失败: %v", err)
				}
				lastW = w
				lastH = h
				displayW = w
				displayH = h
				fmt.Printf("视频尺寸: %dx%d\n", w, h)
			}

			if err := sdl.RenderYUV(nv12); err != nil {
				log.Printf("渲染失败: %v", err)
			}
			frameCount++
			if frameCount%100 == 0 {
				fmt.Printf("已渲染 %d 帧\n", frameCount)
			}
		}
	}
}
