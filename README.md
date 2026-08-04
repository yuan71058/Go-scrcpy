# Go-scrcpy

Go 语言实现的 scrcpy 客户端库，用于通过 [scrcpy](https://github.com/Genymobile/scrcpy) 协议连接和控制 Android 设备。

## 功能特性

- 完整支持 scrcpy v4.1 协议（22 种控制消息 + ws-scrcpy 扩展）
- 多设备同时接入控制
- 视频流 (H.264, H.265, AV1, VP8, VP9)
- 音频流 (Opus, AAC, FLAC, PCM)
- 触摸、键盘、滚动、鼠标输入注入
- 剪贴板双向同步
- 文件推送和 APK 安装
- 屏幕录制 (MP4/MKV) 和截图 (PNG/JPEG)
- 摄像头控制 (闪光灯、变焦)
- UHID 虚拟 HID 设备支持
- 自动设备发现和跟踪

## 环境要求

- Go >= 1.22
- ADB 已安装并在 PATH 中
- Android 设备已开启 USB 调试
- scrcpy-server.jar（单独下载或自行编译）

## 安装

```bash
go get github.com/yuan71058/go-scrcpy
```

## 快速开始

### 单设备连接

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
    opts := scrcpy.DefaultOptions()
    opts.Server.VideoBitRate = 4000000

    client := scrcpy.New("DEVICE_SERIAL", opts)

    ctx := context.Background()
    if err := client.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    for frame := range client.VideoStream() {
        fmt.Printf("帧: PTS=%d, 大小=%d\n", frame.PTS, len(frame.Data))
    }
}
```

### 投屏窗口 (SDL3)

使用 SDL3 渲染视频并支持鼠标/键盘输入：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "runtime"

    "github.com/yuan71058/go-scrcpy/pkg/input"
    "github.com/yuan71058/go-scrcpy/pkg/render"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
    opts := scrcpy.DefaultOptions()
    opts.Server.VideoBitRate = 4000000
    opts.Server.Audio = false
    opts.Server.Control = true

    client := scrcpy.New("DEVICE_SERIAL", opts, 27183)
    ctx := context.Background()
    if err := client.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    hs := client.Handshake().(*protocol.Handshake)
    displayW := int(hs.GetDisplayWidth())
    displayH := int(hs.GetDisplayHeight())

    // 必须在 SDL 操作前锁定 OS 线程（Windows 要求）
    render.InitSDL()
    runtime.LockOSThread()

    // 解码 goroutine
    go func() {
        decoder, _ := render.NewH264Decoder()
        decoder.SetSize(displayW, displayH)
        for frame := range client.VideoStream() {
            yuv, w, h, _ := decoder.Decode(frame.Data)
            if yuv != nil {
                // 渲染...
            }
        }
    }()

    // 主线程事件循环
    for {
        evType, evData := render.PollEvent()
        // 处理事件...
    }
}
```

完整示例见 [_examples/mirror](_examples/mirror/)。

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
    adbClient := adb.NewClient("adb")
    mc := scrcpy.NewMulti(adbClient)
    defer mc.Close()

    // 监听设备上下线
    watcher := scrcpy.WatchDevices(adbClient,
        func(serial string) {
            fmt.Printf("设备上线: %s\n", serial)
            mc.Add(serial, scrcpy.DefaultOptions())
        },
        func(serial string) {
            fmt.Printf("设备离线: %s\n", serial)
            mc.Remove(serial)
        },
    )

    ctx := context.Background()
    watcher.Start(ctx)
}
```

## API 参考

### Client 单设备客户端

```go
// 创建客户端
client := scrcpy.New(serial, opts)

// 启动连接
client.Start(ctx)

// 获取视频流
for frame := range client.VideoStream() {
    // 处理视频帧
}

// 获取音频流
for pkt := range client.AudioStream() {
    // 处理音频数据
}

// 触摸输入
client.TouchDown(pointerID, x, y, screenW, screenH)
client.TouchMove(pointerID, x, y, screenW, screenH)
client.TouchUp(pointerID, x, y, screenW, screenH)

// 鼠标输入
client.MouseDown(x, y, screenW, screenH, button)
client.MouseMove(x, y, screenW, screenH, buttons)
client.MouseUp(x, y, screenW, screenH, button)

// 滚动
client.Scroll(x, y, screenW, screenH, hscroll, vscroll)

// 按键输入
client.KeyDown(keycode, meta)
client.KeyUp(keycode, meta)
client.KeyPress(keycode, meta)
client.Text("Hello")

// 快捷按键
client.Home()
client.Back()
client.Power()
client.VolumeUp()
client.VolumeDown()
client.Menu()
client.Enter()
client.Delete()
client.Tab()
client.Escape()
client.Space()

// 面板控制
client.ExpandNotificationPanel()
client.ExpandSettingsPanel()
client.CollapsePanels()
client.RotateDevice()

// 显示控制
client.SetDisplayPower(on)
client.PowerOn()
client.PowerOff()
client.ResizeDisplay(width, height)

// 剪贴板
client.SetClipboard(text, paste, sequence)
client.GetClipboard()

// 应用控制
client.StartApp(packageName)
client.ScanFile(path)

// 摄像头控制
client.CameraSetTorch(on)
client.CameraZoomIn()
client.CameraZoomOut()

// UHID 设备
client.UHIDCreate(id, vendorID, productID, name, reportDesc)
client.UHIDInput(id, data)
client.UHIDDestroy(id)

// 关闭
client.Close()
```

### MultiClient 多设备管理器

```go
mc := scrcpy.NewMulti(adbClient)

// 添加设备
mc.Add(serial, opts)

// 移除设备
mc.Remove(serial)

// 获取设备
client, exists := mc.Get(serial)

// 列举设备
clients := mc.List()

// 广播到所有设备
mc.Broadcast(msg)

// 遍历设备
mc.ForEach(func(serial string, client *scrcpy.Client) {
    // 处理设备
})

// 合并所有设备的视频流
for vfs := range mc.VideoStreamAll() {
    fmt.Printf("设备 %s: 帧大小 %d\n", vfs.Serial, len(vfs.Frame.Data))
}
```

### Options 配置选项

```go
opts := scrcpy.DefaultOptions()
opts.Server.VideoCodec = "h264"
opts.Server.AudioCodec = "opus"
opts.Server.VideoBitRate = 8000000
opts.Server.MaxFps = 30
opts.Server.MaxSize = 1920
opts.Server.Audio = true
opts.Server.Control = true

// 链式调用
opts := scrcpy.DefaultOptions().
    WithVideoCodec("h265").
    WithMaxFps(60).
    WithBitrate(12000000)
```

## 日志级别

每个模块独立控制日志级别：

```go
import (
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/protocol"
    "github.com/yuan71058/go-scrcpy/pkg/video"
    "github.com/yuan71058/go-scrcpy/pkg/audio"
    "github.com/yuan71058/go-scrcpy/pkg/input"
    "github.com/yuan71058/go-scrcpy/pkg/control"
    "github.com/yuan71058/go-scrcpy/pkg/record"
    "github.com/yuan71058/go-scrcpy/pkg/server"
    "github.com/yuan71058/go-scrcpy/pkg/transport"
)

// 日志级别: LogLevelNone, LogLevelError, LogLevelInfo, LogLevelDebug
scrcpy.SetLogLevel(scrcpy.LogLevelDebug)
adb.SetLogLevel(adb.LogLevelInfo)
protocol.SetLogLevel(protocol.LogLevelError)
// ... 其他模块
```

## 项目结构

```
go-scrcpy/
├── cmd/go-scrcpy/        # CLI 示例
├── pkg/
│   ├── types/            # 公共类型
│   ├── adb/              # ADB 客户端
│   ├── server/           # Server 启动器
│   ├── transport/        # Socket 连接
│   ├── protocol/         # 二进制协议
│   ├── video/            # 视频解码
│   ├── audio/            # 音频解码
│   ├── input/            # 输入事件
│   ├── control/          # 剪贴板和文件推送
│   ├── record/           # 录制和截图
│   └── scrcpy/           # 核心 API
├── _examples/            # 使用示例
│   ├── mirror/           # SDL3 投屏窗口示例
│   ├── single/           # 单设备示例
│   ├── multi/            # 多设备示例
│   ├── control/          # 控制示例
│   └── record/           # 录制示例
└── docs/
    ├── DEVELOPMENT.md    # 开发计划
    └── PROTOCOL.md       # 协议规范
```

## 许可证

MIT
