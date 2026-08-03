// 投屏示例
// 将 scrcpy 视频流 pipe 到 ffplay 实现实时投屏
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func findFFplay() string {
	// 优先使用同目录下的 ffplay
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	if runtime.GOOS == "windows" {
		p := filepath.Join(dir, "ffplay.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		// data 目录
		p = filepath.Join(dir, "..", "data", "ffplay.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		// 开发目录
		p = "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\ffplay.exe"
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "ffplay"
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	adbClient := adb.NewClient("adb")

	// 查找设备
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

	// 启动 scrcpy
	opts := scrcpy.DefaultOptions()
	opts.Server.VideoBitRate = 4000000
	opts.Server.Audio = false
	opts.Server.Control = true
	opts.LocalJAR = filepath.Join(filepath.Dir(os.Args[0]), "..", "data", "scrcpy-server.jar")
	if _, err := os.Stat(opts.LocalJAR); err != nil {
		opts.LocalJAR = "E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\scrcpy-server.jar"
	}

	client := scrcpy.New(serial, opts, 0)
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动 scrcpy 失败: %v", err)
	}

	h := client.Handshake().(*protocol.Handshake)
	fmt.Printf("已连接: %s, %dx%d\n", h.GetDeviceName(), h.GetDisplayWidth(), h.GetDisplayHeight())

	// 启动 ffplay
	ffplayPath := findFFplay()
	fmt.Printf("ffplay: %s\n", ffplayPath)

	title := fmt.Sprintf("scrcpy - %s", h.GetDeviceName())
	cmd := exec.Command(ffplayPath,
		"-loglevel", "warning",
		"-framedrop",
		"-autoexit",
		"-f", "h264",
		"-i", "pipe:0",
		"-window_title", title,
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("创建管道失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动 ffplay 失败: %v", err)
	}
	fmt.Println("投屏中... 按 Ctrl+C 退出")

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 写帧到 ffplay
	go func() {
		for frame := range client.VideoStream() {
			if _, err := stdin.Write(frame.Data); err != nil {
				fmt.Printf("写入失败: %v\n", err)
				return
			}
		}
	}()

	// 等待退出
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
