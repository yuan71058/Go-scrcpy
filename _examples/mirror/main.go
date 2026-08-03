// 投屏示例
// 将 scrcpy 视频流 pipe 到 ffplay 实现实时投屏
// 需要安装 ffplay: sudo apt install ffmpeg (Linux) / choco install ffmpeg (Windows)
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
	scrcpy.SetLogLevel(scrcpy.LogLevelInfo)
	adb.SetLogLevel(adb.LogLevelInfo)

	adbClient := adb.NewClient("adb")

	// 查找设备
	devices, err := adbClient.ListDevices(context.Background())
	if err != nil {
		log.Fatalf("列举设备失败: %v", err)
	}
	if len(devices) == 0 {
		log.Fatal("未找到设备")
	}
	serial := devices[0].Serial
	fmt.Printf("使用设备: %s\n", serial)

	// 启动 scrcpy 客户端
	opts := scrcpy.DefaultOptions()
	opts.Server.VideoBitRate = 4000000
	opts.Server.Audio = false
	opts.Server.Control = true
	opts.LocalJAR = "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\scrcpy-server.jar"

	client := scrcpy.New(serial, opts, 0)

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动 scrcpy 失败: %v", err)
	}

	handshake := client.Handshake().(*protocol.Handshake)
	fmt.Printf("scrcpy 已连接 [%s]: %s, %dx%d\n",
		serial, handshake.GetDeviceName(),
		handshake.GetDisplayWidth(), handshake.GetDisplayHeight())

	// 启动 ffplay
	// -f h264: 输入格式为原始 H264
	// -i pipe:0: 从 stdin 读取
	// -framerate 0: 自动检测帧率
	// 启动 ffplay (使用 data 目录下的 ffplay.exe)
_ffplay := "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\ffplay.exe"
	windowTitle := fmt.Sprintf("scrcpy - %s (%s)", serial, handshake.GetDeviceName())
	cmd := exec.Command(_ffplay,
		"-f", "h264",
		"-i", "pipe:0",
		"-framerate", "0",
		"-analyzeduration", "1000000",
		"-probesize", "32768",
		"-window_title", windowTitle,
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("创建 ffplay stdin 管道失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动 ffplay 失败: %v (请确认 %s 存在)", err, _ffplay)
	}
	fmt.Println("ffplay 已启动，正在投屏...")

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 读取视频帧并写入 ffplay
	done := make(chan struct{})
	go func() {
		defer close(done)
		frameCount := 0
		lastReport := time.Now()

		for frame := range client.VideoStream() {
			// 写入 H264 数据到 ffplay
			if _, err := stdin.Write(frame.Data); err != nil {
				fmt.Printf("写入 ffplay 失败: %v\n", err)
				return
			}

			frameCount++
			if time.Since(lastReport) >= time.Second {
				fmt.Printf("投屏中: %d fps, 帧大小: %d bytes\n", frameCount, len(frame.Data))
				frameCount = 0
				lastReport = time.Now()
			}
		}
	}()

	// 等待退出
	select {
	case <-sigChan:
		fmt.Println("\n正在关闭...")
	case <-done:
		fmt.Println("视频流结束")
	}

	// 清理
	stdin.Close()
	cmd.Process.Kill()
	cmd.Wait()
	client.Close()
	fmt.Println("已关闭")
}
