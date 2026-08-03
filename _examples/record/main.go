// 录制和截图示例
// 演示如何录制屏幕和截取截图
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/record"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
	// 设置日志级别
	scrcpy.SetLogLevel(scrcpy.LogLevelInfo)
	adb.SetLogLevel(adb.LogLevelInfo)

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <serial>")
		os.Exit(1)
	}
	serial := os.Args[1]

	// 创建客户端
	adbClient := adb.NewClient("adb")
	opts := scrcpy.DefaultOptions()
	opts.Server.Audio = false
	opts.LocalJAR = "../../data/scrcpy-server.jar"

	client := scrcpy.New(serial, opts)

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

	fmt.Println("连接成功！")

	// 创建录制器
	recorder, err := record.NewRecorderFromFile("recording.mp4", "mp4")
	if err != nil {
		log.Fatalf("创建录制器失败: %v", err)
	}
	defer recorder.Close()

	// 创建截图器
	screenshot := record.NewScreenshot(1920, 1080)

	// 启动视频录制
	go func() {
		frameCount := 0
		for frame := range client.VideoStream() {
			// 写入录制
			if err := recorder.WriteVideo(frame.Data, frame.PTS, frame.Key); err != nil {
				log.Printf("写入视频失败: %v", err)
				continue
			}

			frameCount++
			if frameCount%100 == 0 {
				fmt.Printf("已录制 %d 帧\n", frameCount)
			}
		}
		fmt.Printf("录制完成，共 %d 帧\n", frameCount)
	}()

	// 5 秒后截取截图
	go func() {
		time.Sleep(5 * time.Second)

		fmt.Println("截取截图...")
		data, err := screenshot.Capture(nil, "png")
		if err != nil {
			log.Printf("截取截图失败: %v", err)
			return
		}

		if err := screenshot.SaveToFile(data, "screenshot.png", "png"); err != nil {
			log.Printf("保存截图失败: %v", err)
			return
		}

		fmt.Println("截图已保存: screenshot.png")
	}()

	// 等待退出信号
	fmt.Println("按 Ctrl+C 停止录制...")
	<-sigChan

	fmt.Println("正在保存录制...")
	if err := recorder.Close(); err != nil {
		log.Printf("保存录制失败: %v", err)
	}

	fmt.Println("录制已保存: recording.mp4")
}
