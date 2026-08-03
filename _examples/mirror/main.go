// 投屏示例
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

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

	opts := scrcpy.DefaultOptions()
	opts.Server.VideoBitRate = 4000000
	opts.Server.Audio = false
	opts.Server.Control = true
	opts.LocalJAR = "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\scrcpy-server.jar"

	client := scrcpy.New(serial, opts, 0)
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动 scrcpy 失败: %v", err)
	}

	h := client.Handshake().(*protocol.Handshake)
	fmt.Printf("已连接: %s, %dx%d\n", h.GetDeviceName(), h.GetDisplayWidth(), h.GetDisplayHeight())

	// 启动 ffplay
	cmd := exec.Command("E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\ffplay.exe",
		"-loglevel", "warning",
		"-sync", "video",
		"-framedrop",
		"-f", "h264",
		"-i", "pipe:0",
		"-window_title", fmt.Sprintf("scrcpy - %s", h.GetDeviceName()),
	)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("创建管道失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动 ffplay 失败: %v", err)
	}
	fmt.Println("投屏中... 按 Ctrl+C 退出")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		count := 0
		for frame := range client.VideoStream() {
			if _, err := stdin.Write(frame.Data); err != nil {
				fmt.Printf("写入失败: %v\n", err)
				return
			}
			count++
			if count <= 3 {
				// 打印前3帧的前16字节，验证 H264 数据
				n := 16
				if len(frame.Data) < n {
					n = len(frame.Data)
				}
				fmt.Printf("帧%d: %d bytes, 首字节: %x, config=%v key=%v\n",
					count, len(frame.Data), frame.Data[:n], frame.Config, frame.Key)
			}
			if count%100 == 0 {
				fmt.Printf("已写入 %d 帧\n", count)
			}
		}
		fmt.Printf("视频流结束，共 %d 帧\n", count)
	}()

	select {
	case <-sigChan:
	case <-time.After(30 * time.Minute):
	}

	fmt.Println("\n关闭中...")
	stdin.Close()
	cmd.Process.Kill()
	cmd.Wait()
	client.Close()
}
