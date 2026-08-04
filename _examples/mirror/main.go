// 投屏示例（使用高精度 Display API）
// 演示如何使用 scrcpy.NewDisplay 快速启动投屏窗口
package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
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

	rootCmd := exec.Command("adb", "-s", serial, "root")
	rootCmd.CombinedOutput()

	opts := scrcpy.DefaultOptions()
	opts.Server.VideoBitRate = 4000000
	opts.Server.Audio = false
	opts.Server.Control = true
	opts.Server.VideoCodec = "h264"

	fmt.Println("创建投屏窗口...")
	display, err := scrcpy.NewDisplay(serial, opts)
	if err != nil {
		log.Fatalf("创建投屏失败: %v", err)
	}
	defer display.Close()

	fmt.Println("投屏窗口已启动，点击窗口进行交互")
	display.Run() // 阻塞直到窗口关闭
	fmt.Println("投屏结束")
}