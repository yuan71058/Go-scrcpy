// 投屏示例
// 通过命名管道将 scrcpy H264 流送到 ffplay 实现实时投屏
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
	"unsafe"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

const (
	ffplayPath  = `E:\SRC\Scrcpy\src\Go-scrcpy\data\ffplay.exe`
	pipeName    = `\\.\pipe\scrcpy-live`
	pipeBufSize = 1024 * 1024 // 1MB
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procCreateNamedW = kernel32.NewProc("CreateNamedPipeW")
	procConnectNamed = kernel32.NewProc("ConnectNamedPipe")
)

func createNamedPipe() (syscall.Handle, error) {
	name, _ := syscall.UTF16PtrFromString(pipeName)
	h, _, err := procCreateNamedW.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(0x00000003), // PIPE_ACCESS_DUPLEX | FILE_FLAG_OVERLAPPED
		uintptr(0x00000000), // PIPE_TYPE_BYTE
		1,                   // max instances
		uintptr(pipeBufSize),
		uintptr(pipeBufSize),
		0,
		0,
	)
	if h == uintptr(syscall.InvalidHandle) {
		return 0, fmt.Errorf("CreateNamedPipe 失败: %v", err)
	}
	return syscall.Handle(h), nil
}

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

	// 创建命名管道
	pipe, err := createNamedPipe()
	if err != nil {
		log.Fatalf("创建命名管道失败: %v", err)
	}
	defer syscall.CloseHandle(pipe)

	fmt.Printf("命名管道: %s\n", pipeName)
	fmt.Println("等待 ffplay 连接...")

	// 启动 ffplay 连接命名管道
	title := fmt.Sprintf("scrcpy - %s", h.GetDeviceName())
	cmd := exec.Command(ffplayPath,
		"-sync", "video",
		"-framedrop",
		"-window_title", title,
		pipeName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("启动 ffplay 失败: %v", err)
	}

	// 等待 ffplay 连接管道
	procConnectNamed.Call(uintptr(pipe), 0)
	fmt.Println("ffplay 已连接，开始投屏... 按 Ctrl+C 退出")

	// 同时保存文件
	h264File, _ := os.Create("E:\\SRC\\Scrcpy\\src\\Go-scrcpy\\data\\live.h264")
	defer h264File.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		count := 0
		buf := make([]byte, 0, 64*1024)
		for frame := range client.VideoStream() {
			h264File.Write(frame.Data)

			// 写入命名管道
			written := 0
			for written < len(frame.Data) {
				n, err := syscall.Write(pipe, frame.Data[written:])
				if err != nil {
					fmt.Printf("写入管道失败: %v\n", err)
					return
				}
				written += n
			}

			count++
			if count%100 == 0 {
				fmt.Printf("已投屏 %d 帧\n", count)
			}
			_ = buf
		}
		fmt.Printf("视频流结束，共 %d 帧\n", count)
	}()

	select {
	case <-sigChan:
	case <-time.After(30 * time.Minute):
	}

	fmt.Println("\n关闭中...")
	syscall.CloseHandle(pipe)
	cmd.Process.Kill()
	cmd.Wait()
	client.Close()
	fmt.Println("H264 已保存到 data/live.h264")
}
