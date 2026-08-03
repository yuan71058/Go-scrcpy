// 投屏示例 - 直接调用 scrcpy 实现投屏+控制
// scrcpy 自带: H264解码、SDL2渲染、触控/按键输入、快捷键
// 快捷键说明:
//   右键: 返回 | 双击右键: Home | 三击右键: 最近任务
//   鼠标中键: Home | Ctrl+C: 复制 | Ctrl+V: 粘贴
//   滚轮: 音量调节 | Ctrl+Shift+O: 屏幕开关
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
)

const scrcpyPath = `D:\PATH\scrcpy\scrcpy.exe`

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

	// 使用系统已安装的 scrcpy (自带 SDL2 + FFmpeg 解码)
	cmd := exec.Command(scrcpyPath,
		"--serial", serial,
		"--video-bit-rate", "4M",
		"--max-size", "0",
		"--no-audio",
		"--window-title", fmt.Sprintf("scrcpy - %s", serial),
		// 控制相关
		"--keyboard=uhid",       // UHID 键盘（支持所有按键）
		"--mouse=uhid",          // UHID 鼠标（支持所有按键）
		"--show-touches",        // 显示触摸点
		"--no-video-codec-lock", // 不锁定编码器
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("启动投屏...")
	fmt.Println("控制说明:")
	fmt.Println("  鼠标点击: 触摸")
	fmt.Println("  右键: 返回 | 双击右键: Home | 三击右键: 最近任务")
	fmt.Println("  鼠标中键: Home | 滚轮: 音量/页面滚动")
	fmt.Println("  Ctrl+C: 复制 | Ctrl+V: 粘贴 | Ctrl+Shift+O: 屏幕开关")
	fmt.Println("  Ctrl+F: 全屏 | Ctrl+G: 1:1 | Ctrl+X: 剪切")
	fmt.Println("  Ctrl+Shift+C: 通知栏 | Ctrl+Shift+N: 设置面板")

	if err := cmd.Run(); err != nil {
		log.Fatalf("scrcpy 退出: %v", err)
	}

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
