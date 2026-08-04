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
	scrcpy.SetLogLevel(scrcpy.LogLevelInfo)
	adb.SetLogLevel(adb.LogLevelInfo)

	adbClient := adb.NewClient("adb")
	mc := scrcpy.NewMulti(adbClient)
	defer mc.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	watcher := scrcpy.WatchDevices(adbClient,
		func(serial string) {
			fmt.Printf("设备上线: %s\n", serial)

			opts := scrcpy.DefaultOptions()
			opts.Server.VideoBitRate = 2000000
			opts.Server.Audio = false
			opts.Server.Control = true
			opts.LocalJAR = "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\scrcpy-server.jar"

			client, err := mc.Add(serial, opts)
			if err != nil {
				log.Printf("接入设备 %s 失败: %v", serial, err)
				return
			}

			// 统计帧率并测试控制功能
			go func() {
				frameCount := 0
				lastReport := time.Now()

				for frame := range client.VideoStream() {
					frameCount++

					// 每秒报告帧率
					if time.Since(lastReport) >= time.Second {
						fps := float64(frameCount) / time.Since(lastReport).Seconds()
						fmt.Printf("[%s] 帧率: %.1f fps, 最近帧大小: %d bytes, config=%v key=%v\n",
							serial, fps, len(frame.Data), frame.Config, frame.Key)
						frameCount = 0
						lastReport = time.Now()
					}
				}
				fmt.Printf("[%s] 视频流结束\n", serial)
			}()
		},
		func(serial string) {
			fmt.Printf("设备离线: %s\n", serial)
			mc.Remove(serial)
		},
	)

	ctx := context.Background()
	if err := watcher.Start(ctx); err != nil {
		log.Fatalf("启动设备跟踪失败: %v", err)
	}

	fmt.Println("按 Ctrl+C 退出...")
	<-sigChan
	fmt.Println("\n正在关闭...")
}