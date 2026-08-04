// 单设备连接示例
// 演示如何连接单个 Android 设备并接收视频流
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
	// 设置日志级别
	scrcpy.SetLogLevel(scrcpy.LogLevelInfo)
	adb.SetLogLevel(adb.LogLevelInfo)

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <serial>")
		fmt.Println("示例: go run main.go emulator-5554")
		os.Exit(1)
	}
	serial := os.Args[1]

	// 创建 scrcpy 客户端
	opts := scrcpy.DefaultOptions()
	opts.Server.VideoBitRate = 4000000 // 4Mbps
	opts.Server.Audio = false          // 暂不处理音频
	opts.LocalJAR = "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\scrcpy-server.jar"

	// 使用随机端口
	listenPort := 27183
	client := scrcpy.New(serial, opts, listenPort)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动客户端
	ctx := context.Background()
	fmt.Printf("连接设备 %s...\n", serial)
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
	defer client.Close()

	fmt.Println("连接成功！接收视频帧...")

	// 读取视频帧
	go func() {
		frameCount := 0
		for frame := range client.VideoStream() {
			frameCount++
			if frameCount%100 == 0 {
				fmt.Printf("已接收 %d 帧, PTS=%d, size=%d, config=%v, key=%v\n",
					frameCount, frame.PTS, len(frame.Data), frame.Config, frame.Key)
			}
		}
		fmt.Printf("视频流结束，共接收 %d 帧\n", frameCount)
	}()

	// 模拟发送按键
	go func() {
		// 等待 3 秒后按 HOME 键
		fmt.Println("等待 3 秒...")
		select {
		case <-sigChan:
			return
		case <-func() <-chan struct{} {
			ch := make(chan struct{})
			go func() {
				// 简单等待
				for i := 0; i < 30; i++ {
					select {
					case <-sigChan:
						close(ch)
						return
					default:
					}
				}
				close(ch)
			}()
			return ch
		}():
		}

		fmt.Println("发送 HOME 键...")
		if err := client.Home(); err != nil {
			log.Printf("发送 HOME 键失败: %v", err)
		}
	}()

	// 等待退出信号
	fmt.Println("按 Ctrl+C 退出...")
	<-sigChan
	fmt.Println("\n正在关闭...")
}
