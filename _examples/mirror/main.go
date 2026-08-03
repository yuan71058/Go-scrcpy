// 投屏示例
// 将 scrcpy 视频流 pipe 到 ffplay 实现实时投屏
// 同时保存 H264 文件到 data/live.h264
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

const ffplayPath = `E:\SRC\Scrcpy\src\Go-scrcpy\data\ffplay.exe`

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

	// 保存 H264 文件
	h264File, _ := os.Create("E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\live.h264")
	defer h264File.Close()

	// 启动 ffplay
	cmd := exec.Command(ffplayPath,
		"-sync", "video",
		"-framedrop",
		"-window_title", fmt.Sprintf("scrcpy - %s", h.GetDeviceName()),
		"-i", "pipe:0",
	)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stderr

	stdin, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动 ffplay 失败: %v", err)
	}

	fmt.Println("投屏中... 按 Ctrl+C 退出")
	fmt.Println("如果看不到窗口，请运行: ffplay -sync video data/live.h264")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		count := 0
		for frame := range client.VideoStream() {
			h264File.Write(frame.Data)
			stdin.Write(frame.Data)
			count++
			if count%100 == 0 {
				fmt.Printf("已投屏 %d 帧\n", count)
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
	fmt.Println("H264 已保存到 data/live.h264")
}
