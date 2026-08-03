// 投屏示例 - 使用项目自己的 Go 代码实现投屏+控制
// 通过 pkg/scrcpy 接收视频流，pkg/render 解码渲染
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/render"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

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

	// adb root
	rootCmd := exec.Command("adb", "-s", serial, "root")
	rootCmd.CombinedOutput()

	// scrcpy-server.jar 路径
	// 从 _examples/mirror 向上两级到 project root
	exe, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(exe))
	localJAR := filepath.Join(projectRoot, "data", "scrcpy-server.jar")
	fmt.Printf("JAR: %s\n", localJAR)

	// 创建 scrcpy 客户端
	opts := scrcpy.DefaultOptions()
	opts.LocalJAR = localJAR
	opts.Server.VideoBitRate = 4000000
	opts.Server.Audio = false
	opts.Server.Control = true

	listenPort := 27183
	client := scrcpy.New(serial, opts, listenPort)

	// 启动客户端
	fmt.Println("连接设备...")
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
	defer client.Close()

	// 创建解码器
	decoder, err := render.NewH264Decoder()
	if err != nil {
		log.Fatalf("创建解码器失败: %v", err)
	}
	defer decoder.Close()

	fmt.Println("连接成功！启动投屏窗口...")

	// 创建 SDL 渲染器 (初始尺寸 0,0 会在收到第一帧时调整)
	var sdl *render.SDLRenderer

	frameCount := 0
	var lastW, lastH int

	// 读取视频帧并渲染
	for frame := range client.VideoStream() {
		if !frame.Config {
			// 解码 H264 帧
			yuv, w, h, err := decoder.Decode(frame.Data)
			if err != nil {
				continue
			}

			// 尺寸变化时重建渲染器
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
				fmt.Printf("视频尺寸: %dx%d\n", w, h)
			}

			// YUV -> RGB
			rgb := render.YUV420PtoRGB(yuv, w, h)

			// 渲染
			if err := sdl.RenderYUV(rgb); err != nil {
				log.Printf("渲染失败: %v", err)
			}

			frameCount++
			if frameCount%100 == 0 {
				fmt.Printf("已渲染 %d 帧\n", frameCount)
			}
		}
	}

	if sdl != nil {
		sdl.Destroy()
	}
	fmt.Printf("投屏结束，共渲染 %d 帧\n", frameCount)
}
