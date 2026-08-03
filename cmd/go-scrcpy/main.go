// go-scrcpy 是 Go-scrcpy 库的 CLI 示例程序
// 演示如何使用库连接 Android 设备并接收视频流
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
	fmt.Println("Go-scrcpy CLI 示例")
	fmt.Println("==================")

	// 设置日志级别
	scrcpy.SetLogLevel(scrcpy.LogLevelInfo)
	adb.SetLogLevel(adb.LogLevelInfo)

	// 创建 ADB 客户端
	adbClient := adb.NewClient("adb")

	// 列举设备
	ctx := context.Background()
	devices, err := adbClient.ListDevices(ctx)
	if err != nil {
		fmt.Printf("列举设备失败: %v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("未发现设备，请连接 Android 设备")
		os.Exit(1)
	}

	fmt.Printf("发现 %d 个设备:\n", len(devices))
	for i, d := range devices {
		fmt.Printf("  [%d] %s (%s) - %s\n", i+1, d.Serial, d.State, d.Model)
	}

	// 使用第一个设备
	serial := devices[0].Serial
	fmt.Printf("\n使用设备: %s\n", serial)

	// 创建客户端
	opts := scrcpy.DefaultOptions()
	opts.Server.VideoBitRate = 4000000 // 4Mbps
	opts.Server.Audio = false          // 暂不处理音频
	opts.Server.SendFrameMeta = true

	listenPort := 27183
	client := scrcpy.New(serial, opts, listenPort)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动客户端
	fmt.Println("启动 scrcpy-server...")
	if err := client.Start(ctx); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("连接成功！接收视频帧...")

	// 读取视频帧
	go func() {
		frameCount := 0
		for frame := range client.VideoStream() {
			frameCount++
			if frameCount%100 == 0 {
				fmt.Printf("已接收 %d 帧, 最新帧: PTS=%d, size=%d, config=%v, key=%v\n",
					frameCount, frame.PTS, len(frame.Data), frame.Config, frame.Key)
			}
		}
		fmt.Printf("视频流结束，共接收 %d 帧\n", frameCount)
	}()

	// 等待退出信号
	fmt.Println("按 Ctrl+C 退出...")
	<-sigChan
	fmt.Println("\n正在关闭...")
}
