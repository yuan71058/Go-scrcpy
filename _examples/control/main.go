// 远程控制示例
// 演示如何使用各种输入控制功能
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
	"github.com/yuan71058/go-scrcpy/pkg/input"
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

	listenPort := 27184
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

	fmt.Println("连接成功！")

	// 启动视频接收
	go func() {
		for frame := range client.VideoStream() {
			_ = frame
		}
	}()

	// 演示各种控制功能
	go func() {
		time.Sleep(2 * time.Second)

		fmt.Println("=== 演示控制功能 ===")

		// 1. 按 HOME 键
		fmt.Println("1. 按 HOME 键")
		client.Home()
		time.Sleep(1 * time.Second)

		// 2. 按返回键
		fmt.Println("2. 按返回键")
		client.Back()
		time.Sleep(1 * time.Second)

		// 3. 音量加
		fmt.Println("3. 音量加")
		client.VolumeUp()
		time.Sleep(1 * time.Second)

		// 4. 音量减
		fmt.Println("4. 音量减")
		client.VolumeDown()
		time.Sleep(1 * time.Second)

		// 5. 展开通知栏
		fmt.Println("5. 展开通知栏")
		client.ExpandNotificationPanel()
		time.Sleep(2 * time.Second)

		// 6. 收起面板
		fmt.Println("6. 收起面板")
		client.CollapsePanels()
		time.Sleep(1 * time.Second)

		// 7. 文本输入
		fmt.Println("7. 文本输入")
		client.Text("Hello from Go-scrcpy!")
		time.Sleep(1 * time.Second)

		// 8. 启动应用
		fmt.Println("8. 启动设置应用")
		client.StartApp("com.android.settings")
		time.Sleep(2 * time.Second)

		// 9. 旋转设备
		fmt.Println("9. 旋转设备")
		client.RotateDevice()
		time.Sleep(2 * time.Second)

		// 10. 设置剪贴板
		fmt.Println("10. 设置剪贴板")
		client.SetClipboard("测试文本", true, 1)
		time.Sleep(1 * time.Second)

		fmt.Println("=== 演示完成 ===")
	}()

	// 等待退出信号
	fmt.Println("按 Ctrl+C 退出...")
	<-sigChan
	fmt.Println("\n正在关闭...")
}
