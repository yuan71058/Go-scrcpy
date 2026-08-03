// 多设备并发示例
// 演示如何同时连接多个 Android 设备
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
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
	// 设置日志级别
	scrcpy.SetLogLevel(scrcpy.LogLevelInfo)
	adb.SetLogLevel(adb.LogLevelInfo)

	// 创建 ADB 客户端
	adbClient := adb.NewClient("adb")

	// 创建多设备管理器
	mc := scrcpy.NewMulti(adbClient)
	defer mc.Close()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 监听设备上下线
	watcher := scrcpy.WatchDevices(adbClient,
		// 设备上线
		func(serial string) {
			fmt.Printf("设备上线: %s\n", serial)

			// 为新设备创建客户端
			opts := scrcpy.DefaultOptions()
			opts.Server.VideoBitRate = 2000000 // 2Mbps
			opts.Server.Audio = false

			client, err := mc.Add(serial, opts)
			if err != nil {
				log.Printf("接入设备 %s 失败: %v", serial, err)
				return
			}

			// 启动视频接收
			go func() {
				frameCount := 0
				for range client.VideoStream() {
					frameCount++
					if frameCount%100 == 0 {
						fmt.Printf("[%s] 已接收 %d 帧\n", serial, frameCount)
					}
				}
				fmt.Printf("[%s] 视频流结束\n", serial)
			}()
		},
		// 设备离线
		func(serial string) {
			fmt.Printf("设备离线: %s\n", serial)
			mc.Remove(serial)
		},
	)

	// 启动设备跟踪
	ctx := context.Background()
	if err := watcher.Start(ctx); err != nil {
		log.Fatalf("启动设备跟踪失败: %v", err)
	}

	// 定期广播
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-sigChan:
				return
			case <-ticker.C:
				count := mc.Count()
				if count > 0 {
					fmt.Printf("当前连接 %d 个设备\n", count)
					// 向所有设备发送通知栏展开
					mc.ForEach(func(serial string, client *scrcpy.Client) {
						fmt.Printf("  [%s] 发送展开通知栏\n", serial)
						client.ExpandNotificationPanel()
					})
				}
			}
		}
	}()

	// 等待退出信号
	fmt.Println("按 Ctrl+C 退出...")
	<-sigChan
	fmt.Println("\n正在关闭...")
}
