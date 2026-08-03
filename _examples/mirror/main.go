// 投屏示例
package main

import (
	"context"
	"fmt"
	"io"
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

	// io.Pipe: 写端给 Go，读端给 ffplay
	pr, pw := io.Pipe()

	// 启动 ffplay 从 stdin 读取 H264
	title := fmt.Sprintf("scrcpy - %s", h.GetDeviceName())
	cmd := exec.Command(ffplayPath,
		"-loglevel", "warning",
		"-sync", "video",
		"-framedrop",
		"-probesize", "32",
		"-analyzeduration", "0",
		"-f", "h264",
		"-i", "pipe:0",
		"-window_title", title,
	)
	cmd.Stdin = pr
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动 ffplay 失败: %v", err)
	}
	fmt.Println("投屏中... 按 Ctrl+C 退出")

	// 同时保存文件
	h264File, _ := os.Create("E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\live.h264")
	defer h264File.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		count := 0
		for frame := range client.VideoStream() {
			h264File.Write(frame.Data)
			pw.Write(frame.Data)
			count++
			if count%100 == 0 {
				fmt.Printf("已投屏 %d 帧\n", count)
			}
		}
		fmt.Printf("视频流结束，共 %d 帧\n", count)
		pw.Close()
	}()

	select {
	case <-sigChan:
	case <-time.After(30 * time.Minute):
	}

	fmt.Println("\n关闭中...")
	pw.Close()
	cmd.Process.Kill()
	cmd.Wait()
	client.Close()
}
