<div align="center">

# Go-scrcpy

**Go 语言实现的 Android 设备投屏与控制库**

基于 [scrcpy-server](https://github.com/Genymobile/scrcpy) 协议，提供完整的 Android 设备屏幕镜像、远程控制、文件管理能力。

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Protocol](https://img.shields.io/badge/scrcpy--server-v4.1-blueviolet)](https://github.com/Genymobile/scrcpy)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

</div>

---

## 目录

- [概览](#概览)
- [功能特性](#功能特性)
- [架构](#架构)
- [环境要求](#环境要求)
- [安装](#安装)
- [快速开始](#快速开始)
  - [单设备连接](#单设备连接)
  - [多设备并发管理](#多设备并发管理)
  - [投屏窗口 (SDL3 + FFmpeg)](#投屏窗口-sdl3--ffmpeg)
  - [远程控制](#远程控制)
  - [录制与截图](#录制与截图)
- [API 参考](#api-参考)
  - [Client](#client-单设备客户端)
  - [Display](#display-投屏窗口)
  - [MultiClient](#multiclient-多设备管理器)
  - [Options](#options-配置选项)
  - [DeviceWatcher](#devicewatcher-设备自动发现)
  - [日志级别](#日志级别)
- [项目结构](#项目结构)
- [协议支持](#协议支持)
- [许可证](#许可证)

---

## 概览

Go-scrcpy 是一个纯 Go 编写的 scrcpy 客户端库，通过 ADB (Android Debug Bridge) 与 Android 设备建立连接，实现：

- **屏幕镜像** — 实时接收设备屏幕的 H.264/H.265 视频流
- **音频流** — 同步传输设备音频 (Opus/AAC/FLAC/PCM)
- **远程控制** — 触摸、按键、鼠标、滚动、文本注入
- **剪贴板同步** — 双向剪贴板读写
- **文件管理** — APK 安装、文件推送
- **屏幕录制** — MP4/MKV 录制、PNG/JPEG 截图
- **多设备并发** — 同时接入和控制多台设备
- **投屏窗口** — 基于 SDL3 + FFmpeg 的硬件加速渲染窗口

---

## 功能特性

### 核心能力

| 类别 | 功能 | 说明 |
|------|------|------|
| 视频 | H.264 / H.265 / AV1 / VP8 / VP9 | 5 种视频编解码器支持 |
| 音频 | Opus / AAC / FLAC / PCM | 4 种音频编解码器，支持 10+ 种音频源 |
| 输入 | 触摸 / 按键 / 鼠标 / 滚动 / 文本 | 完整 Android 输入事件注入 |
| 控制 | 通知栏 / 设置面板 / 旋转 / 电源 | 系统级快捷控制 |
| 剪贴板 | 双向同步 | 读取和设置设备剪贴板内容 |
| 文件 | APK 安装 / 文件推送 | 通过 ws-scrcpy 扩展协议 |
| 录制 | MP4 / MKV | 视频录制到本地文件 |
| 截图 | PNG / JPEG | 保存设备屏幕截图 |
| 摄像头 | 闪光灯 / 变焦 | 远程摄像头控制 |
| UHID | 虚拟 HID 设备 | 创建、输入、销毁虚拟 HID 设备 |

### 架构特性

- **三通道通信** — 视频、音频、控制独立 Socket 通道，互不阻塞
- **多设备并发** — 每个设备独立 goroutine，支持数十台设备同时接入
- **自动设备发现** — 实时监听设备上下线，自动连接/断开
- **模块化设计** — 各模块独立，日志级别可单独控制
- **纯 Go 核心** — 核心库无 CGo 依赖，投屏渲染可选 FFmpeg

---

## 架构

```
┌──────────────────────────────────────────────────────────────┐
│                     Go-scrcpy Library                         │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                    │
│  │  Client   │  │  Client   │  │  Client   │  ...              │
│  │ (DeviceA) │  │ (DeviceB) │  │ (DeviceC) │                    │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘                    │
│        │              │              │                         │
│  ┌─────┴──────────────┴──────────────┴─────┐                  │
│  │          MultiClient Manager              │                  │
│  └─────────────────┬───────────────────────┘                  │
│                    │                                           │
│  ┌─────────────────┴───────────────────────┐                  │
│  │              DeviceSession                │                  │
│  │  ┌──────────┐  ┌──────────┐  ┌────────┐ │                  │
│  │  │  Video   │  │  Audio   │  │ Control│ │                  │
│  │  │ Channel  │  │ Channel  │  │ Channel│ │                  │
│  │  └────┬─────┘  └────┬─────┘  └───┬────┘ │                  │
│  └───────┼──────────────┼──────────────┼─────┘                  │
│          │              │              │                         │
│  ┌───────┴──────────────┴──────────────┴─────┐                  │
│  │         Transport Layer (Socket)            │                  │
│  └─────────────────┬───────────────────────┘                  │
└────────────────────┼──────────────────────────────────────────┘
                     │
              ADB Reverse Tunnel
                     │
┌────────────────────┼──────────────────────────────────────────┐
│           Android Device (scrcpy-server)                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                    │
│  │  Video   │  │  Audio   │  │  Control │                    │
│  │  Encoder │  │  Encoder │  │  Handler │                    │
│  └──────────┘  └──────────┘  └──────────┘                    │
└──────────────────────────────────────────────────────────────┘
```

通信流程：

1. **ADB 连接** — 通过 `adb push` 部署 scrcpy-server.jar，使用 `adb reverse` 建立端口转发
2. **三通道建立** — 视频、音频、控制三个独立 Socket 连接
3. **握手** — 获取设备名称、显示信息、编解码器列表
4. **流传输** — 视频/音频帧持续推送，控制消息双向通信
5. **资源清理** — 断开连接时自动关闭通道、清理端口转发

---

## 环境要求

| 依赖 | 说明 |
|------|------|
| **Go** | >= 1.22 |
| **ADB** | 已安装并在 PATH 中 |
| **Android 设备** | 已开启 USB 调试 |
| **scrcpy-server.jar** | 需要单独下载（[官方发布页](https://github.com/Genymobile/scrcpy/releases)）或自行编译 |
| **FFmpeg DLL** (可选) | 投屏窗口需要 ffmpeg DLL (Windows) |

scrcpy-server.jar 放置位置（自动搜索）：

- `./data/scrcpy-server.jar`
- 可执行文件同目录下的 `data/scrcpy-server.jar`
- 向上遍历查找 `go.mod` 所在目录的 `data/scrcpy-server.jar`

---

## 安装

```bash
go get github.com/yuan71058/go-scrcpy
```

---

## 快速开始

### 单设备连接

最简单的使用方式 — 连接设备并接收视频流：

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
    opts.Server.VideoBitRate = 4000000 // 4 Mbps

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

完整示例见 [\_examples/single](_examples/single/)。

---

### 多设备并发管理

使用 `MultiClient` 管理多台设备，结合 `DeviceWatcher` 实现自动发现：

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
            client, err := mc.Add(serial, scrcpy.DefaultOptions())
            if err != nil {
                log.Printf("接入设备 %s 失败: %v", serial, err)
                return
            }
            // 每个设备独立 goroutine 处理视频流
            go func() {
                for frame := range client.VideoStream() {
                    _ = frame // 处理视频帧
                }
            }()
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

完整示例见 [\_examples/multi](_examples/multi/)。

---

### 投屏窗口 (SDL3 + FFmpeg)

使用 `Display` API 几行代码即可启动带硬件加速的投屏窗口，支持鼠标和键盘交互：

```go
package main

import (
    "fmt"
    "log"
    "os/exec"

    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
    adbClient := adb.NewClient("adb")
    devices, _ := adbClient.ListDevices(context.Background())
    if len(devices) == 0 {
        log.Fatal("未找到设备")
    }
    serial := devices[0].Serial

    // 某些设备需要 root 权限启动 scrcpy-server
    exec.Command("adb", "-s", serial, "root").CombinedOutput()

    opts := scrcpy.DefaultOptions()
    opts.Server.VideoBitRate = 4000000
    opts.Server.Audio = false
    opts.Server.VideoCodec = "h264"

    display, err := scrcpy.NewDisplay(serial, opts)
    if err != nil {
        log.Fatal(err)
    }
    defer display.Close()

    fmt.Println("投屏窗口已启动，点击窗口进行交互")
    display.Run() // 阻塞直到窗口关闭
}
```

完整示例见 [\_examples/mirror](_examples/mirror/)。

---

### 远程控制

对已连接的设备发送各种控制指令：

```go
// 基础按键
client.Home()
client.Back()
client.Power()
client.VolumeUp()
client.VolumeDown()

// 系统面板
client.ExpandNotificationPanel()
client.ExpandSettingsPanel()
client.CollapsePanels()

// 显示控制
client.SetDisplayPower(true)
client.PowerOn()
client.RotateDevice()

// 应用控制
client.StartApp("com.android.settings")

// 文本输入
client.Text("Hello from Go-scrcpy!")

// 剪贴板操作
client.SetClipboard("text", true, 1) // paste=true
client.GetClipboard()

// 触摸事件
client.TouchDown(pointerID, x, y, screenW, screenH)
client.TouchMove(pointerID, x, y, screenW, screenH)
client.TouchUp(pointerID, x, y, screenW, screenH)

// 鼠标事件
client.MouseDown(x, y, screenW, screenH, button)
client.MouseUp(x, y, screenW, screenH, button)

// 滚动
client.Scroll(x, y, screenW, screenH, hscroll, vscroll)
```

完整示例见 [\_examples/control](_examples/control/)。

---

### 录制与截图

录制设备屏幕到 MP4 文件，或截取 PNG 截图：

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yuan71058/go-scrcpy/pkg/record"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
)

func main() {
    client := scrcpy.New("DEVICE_SERIAL", scrcpy.DefaultOptions())
    ctx := context.Background()
    client.Start(ctx)
    defer client.Close()

    // 创建录制器
    recorder, _ := record.NewRecorderFromFile("recording.mp4", "mp4")
    defer recorder.Close()

    // 创建截图器
    screenshot := record.NewScreenshot(1920, 1080)

    for frame := range client.VideoStream() {
        // 录制视频帧
        recorder.WriteVideo(frame.Data, frame.PTS, frame.Key)

        // 截图
        screenshot.SaveToFile(nil, "screenshot.png", "png")
        break
    }
}
```

完整示例见 [\_examples/record](_examples/record/)。

---

## API 参考

### Client 单设备客户端

`Client` 封装与单个 Android 设备的完整 scrcpy 连接。

```go
// 创建客户端
client := scrcpy.New(serial, opts)

// 启动连接（建立视频/音频/控制三通道）
client.Start(ctx)

// 获取视频流（返回 <-chan *video.DecodedFrame）
for frame := range client.VideoStream() {
    // 处理视频帧
}

// 获取音频流（返回 <-chan *audio.AudioData）
for pkt := range client.AudioStream() {
    // 处理音频数据
}

// 获取设备信息
info := client.DeviceInfo()

// 获取握手数据
hs := client.Handshake()

// 关闭连接
client.Close()
```

#### 输入控制

```go
// 触摸事件
client.TouchDown(pointerID uint64, x, y int32, screenW, screenH uint16)
client.TouchMove(pointerID uint64, x, y int32, screenW, screenH uint16)
client.TouchUp(pointerID uint64, x, y int32, screenW, screenH uint16)

// 鼠标事件
client.MouseDown(x, y int32, screenW, screenH uint16, button int32)
client.MouseMove(x, y int32, screenW, screenH uint16, buttons int32)
client.MouseUp(x, y int32, screenW, screenH uint16, button int32)

// 滚动
client.Scroll(x, y int32, screenW, screenH uint16, hscroll, vscroll int16)

// 按键事件
client.KeyDown(keycode, meta int32) error
client.KeyUp(keycode, meta int32) error
client.KeyPress(keycode, meta int32) error
client.Text(text string) error
```

#### 快捷按键

```go
client.Home() error
client.Back() error
client.Power() error
client.VolumeUp() error
client.VolumeDown() error
client.Menu() error
client.Enter() error
client.Delete() error
client.Tab() error
client.Escape() error
client.Space() error
```

#### 面板控制

```go
client.ExpandNotificationPanel() error
client.ExpandSettingsPanel() error
client.CollapsePanels() error
client.RotateDevice() error
```

#### 显示控制

```go
client.SetDisplayPower(on bool) error
client.PowerOn() error
client.PowerOff() error
client.ResizeDisplay(width, height uint16) error
```

#### 剪贴板

```go
client.SetClipboard(text string, paste bool, sequence uint64) error
client.GetClipboard() error
```

#### 应用控制

```go
client.StartApp(packageName string) error
client.ScanFile(path string) error
```

#### 摄像头控制

```go
client.CameraSetTorch(on bool) error
client.CameraZoomIn() error
client.CameraZoomOut() error
```

#### UHID 虚拟 HID 设备

```go
client.UHIDCreate(id uint16, vendorID, productID uint16, name string, reportDesc []byte) error
client.UHIDInput(id uint16, data []byte) error
client.UHIDDestroy(id uint16) error
```

---

### Display 投屏窗口

`Display` 封装完整的 SDL3 投屏流程：创建窗口、FFmpeg 硬件解码视频、处理输入事件映射。

```go
// 创建投屏窗口
display, err := scrcpy.NewDisplay(serial, opts)

// 运行窗口（阻塞直到窗口关闭）
display.Run()

// 关闭窗口
display.Close()
```

SDL 按键到 Android 按键的映射表定义在 `SDLToAndroidKey` 中，支持字母、数字、功能键的完整映射。

---

### MultiClient 多设备管理器

`MultiClient` 支持同时管理多台 Android 设备，线程安全。

```go
// 创建多设备管理器
mc := scrcpy.NewMulti(adbClient)

// 添加设备
client, err := mc.Add(serial, opts)

// 移除设备
mc.Remove(serial)

// 获取设备
client, exists := mc.Get(serial)

// 列举所有设备
clients := mc.List()

// 获取设备数
count := mc.Count()

// 广播消息到所有设备
mc.Broadcast(msg []byte)

// 遍历所有设备并执行操作
mc.ForEach(func(serial string, client *scrcpy.Client) {
    // 处理设备
})

// 合并所有设备的视频流（返回 channel）
for vfs := range mc.VideoStreamAll() {
    fmt.Printf("设备 %s: 帧大小 %d\n", vfs.Serial, len(vfs.Frame.Data))
}

// 关闭所有设备连接
mc.Close()
```

---

### DeviceWatcher 设备自动发现

`DeviceWatcher` 实时监听 ADB 设备上下线事件，自动触发回调。

```go
watcher := scrcpy.WatchDevices(adbClient,
    func(serial string) {
        // 设备上线回调
    },
    func(serial string) {
        // 设备离线回调
    },
)

watcher.Start(ctx)
```

---

### Options 配置选项

`Options` 支持链式调用和直接属性设置两种方式：

```go
// 直接属性设置
opts := scrcpy.DefaultOptions()
opts.Server.VideoCodec = "h264"
opts.Server.AudioCodec = "opus"
opts.Server.VideoBitRate = 8000000
opts.Server.MaxFps = 30
opts.Server.MaxSize = 1920
opts.Server.Audio = true
opts.Server.Control = true
opts.Server.ClipboardAutosync = true
opts.Server.PowerOn = true
opts.Server.Cleanup = true

// 链式调用
opts := scrcpy.DefaultOptions().
    WithVideoCodec("h265").
    WithMaxFps(60).
    WithBitrate(12000000)
```

#### Server 选项详解

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| **视频** | | | |
| `Video` | bool | true | 启用视频捕获 |
| `VideoCodec` | string | "h264" | 编解码器: h264, h265, av1, vp8, vp9 |
| `VideoSource` | string | "display" | 视频源: display, camera |
| `VideoBitRate` | int | 8000000 | 视频码率 (bps) |
| `MaxFps` | int | 0 (无限制) | 最大帧率 |
| `MaxSize` | int | 0 (无限制) | 最大分辨率 (短边) |
| `Crop` | string | "" | 画面裁剪: "width:height:x:y" |
| `DisplayID` | int | 0 | 显示器 ID |
| **音频** | | | |
| `Audio` | bool | true | 启用音频捕获 |
| `AudioCodec` | string | "opus" | 编解码器: opus, aac, flac, raw |
| `AudioSource` | string | "output" | 音频源: output, mic, playback, voice-call 等 |
| `AudioBitRate` | int | 128000 | 音频码率 (bps) |
| **控制** | | | |
| `Control` | bool | true | 启用控制通道 |
| `ClipboardAutosync` | bool | true | 剪贴板自动同步 |
| `ShowTouches` | bool | false | 显示触摸点 |
| `StayAwake` | bool | false | 保持屏幕常亮 |
| `PowerOn` | bool | true | 启动时亮屏 |
| **高级** | | | |
| `DownsizeOnError` | bool | true | 编码出错时降分辨率 |
| `Cleanup` | bool | true | 退出时清理 |
| `SendFrameMeta` | bool | true | 发送帧元数据 (PTS/flags) |
| `SCID` | int | -1 | 会话客户端 ID |

---

### 日志级别

每个模块独立控制日志级别，支持精细化调试：

```go
import (
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/protocol"
    "github.com/yuan71058/go-scrcpy/pkg/transport"
    "github.com/yuan71058/go-scrcpy/pkg/video"
    "github.com/yuan71058/go-scrcpy/pkg/audio"
    "github.com/yuan71058/go-scrcpy/pkg/input"
    "github.com/yuan71058/go-scrcpy/pkg/control"
    "github.com/yuan71058/go-scrcpy/pkg/record"
    "github.com/yuan71058/go-scrcpy/pkg/server"
)

// 日志级别: LogLevelNone (0), LogLevelError (1), LogLevelInfo (2), LogLevelDebug (3)
scrcpy.SetLogLevel(scrcpy.LogLevelDebug)   // 核心模块调试
adb.SetLogLevel(adb.LogLevelInfo)           // ADB 信息
protocol.SetLogLevel(protocol.LogLevelError) // 协议层只显示错误
server.SetLogLevel(server.LogLevelDebug)    // Server 启动过程调试
video.SetLogLevel(video.LogLevelInfo)       // 视频流信息
audio.SetLogLevel(audio.LogLevelError)      // 音频只显示错误
input.SetLogLevel(input.LogLevelInfo)       // 输入事件信息
control.SetLogLevel(control.LogLevelNone)   // 控制模块关闭日志
record.SetLogLevel(record.LogLevelInfo)     // 录制信息
transport.SetLogLevel(transport.LogLevelError) // 传输层只显示错误
```

---

## 项目结构

```
go-scrcpy/
├── cmd/
│   └── go-scrcpy/
│       └── main.go              # CLI 示例入口
├── pkg/
│   ├── scrcpy/                  # 核心 API (对外接口)
│   │   ├── client.go            # 单设备客户端
│   │   ├── multi.go             # 多设备管理器
│   │   ├── session.go           # 设备会话管理
│   │   ├── display.go           # SDL3 投屏窗口
│   │   ├── options.go           # 启动选项（链式调用）
│   │   └── log.go               # 日志级别控制
│   ├── adb/                     # ADB 命令封装
│   │   ├── client.go            # ADB 客户端 (设备列举、Shell 等)
│   │   ├── device.go            # 设备发现与跟踪
│   │   └── forward.go           # 端口转发管理
│   ├── server/                  # scrcpy-server 生命周期
│   │   ├── launcher.go          # Server 启动/停止
│   │   ├── options.go           # Server 50+ 参数序列化
│   │   └── version.go           # 版本管理
│   ├── transport/               # Socket 通信层
│   │   ├── connection.go        # 三通道连接管理
│   │   ├── reader.go            # 带缓冲的协议读取器
│   │   └── writer.go            # 协议写入器
│   ├── protocol/                # scrcpy 二进制协议
│   │   ├── handshake.go         # 握手解析
│   │   ├── control_msg.go       # 控制消息序列化 (22+2 种)
│   │   ├── device_msg.go        # 设备消息反序列化
│   │   ├── frame_header.go      # 帧元数据解析
│   │   └── codec.go             # 编解码器 ID 常量
│   ├── video/                   # 视频处理
│   │   ├── decoder.go           # 解码器接口
│   │   ├── h264.go              # H.264 解码
│   │   └── nalu.go              # NALU 解析
│   ├── audio/                   # 音频处理
│   │   ├── decoder.go           # 解码器接口
│   │   ├── opus.go              # Opus 解码
│   │   └── pcm.go               # PCM 播放
│   ├── input/                   # 输入事件
│   │   ├── touch.go             # 触摸事件
│   │   ├── keycode.go           # 按键事件
│   │   └── keycode_map.go       # Android 键码映射表
│   ├── control/                 # 高级控制
│   │   ├── clipboard.go         # 剪贴板同步
│   │   └── filepush.go          # 文件推送/APK 安装
│   ├── record/                  # 录制与截图
│   │   ├── recorder.go          # MP4/MKV 录制
│   │   └── screenshot.go        # PNG/JPEG 截图
│   ├── render/                  # 渲染引擎 (SDL3 + FFmpeg)
│   │   ├── renderer.go          # SDL3 窗口渲染
│   │   ├── decoder.go           # FFmpeg 硬件解码
│   │   ├── dll.go               # DLL 动态加载
│   │   └── convert.go           # 色彩空间转换
│   └── types/                   # 公共类型定义
│       ├── device.go            # DeviceInfo, DisplayInfo, ScreenInfo
│       └── codec.go             # VideoCodec, AudioCodec 枚举
├── _examples/                   # 使用示例
│   ├── mirror/                  # SDL3 投屏窗口示例
│   ├── single/                  # 单设备连接示例
│   ├── multi/                   # 多设备并发示例
│   ├── control/                 # 远程控制示例
│   └── record/                  # 录制截图示例
├── data/                        # 运行时依赖
│   └── bin/                     # FFmpeg 二进制文件
├── docs/
│   ├── DEVELOPMENT.md           # 开发计划与设计文档
│   └── PROTOCOL.md              # 二进制协议规范
├── go.mod
└── README.md
```

---

## 协议支持

完整支持 scrcpy v4.1 协议，包括 22 种控制消息 + 2 种 ws-scrcpy 扩展。

### 控制消息 (Client → Server)

| ID | 消息 | 说明 |
|----|------|------|
| 0 | INJECT_KEYCODE | 注入按键事件 |
| 1 | INJECT_TEXT | 注入文本 |
| 2 | INJECT_TOUCH_EVENT | 注入触摸事件 |
| 3 | INJECT_SCROLL_EVENT | 注入滚动事件 |
| 4 | BACK_OR_SCREEN_ON | 返回或亮屏 |
| 5 | EXPAND_NOTIFICATION_PANEL | 展开通知栏 |
| 6 | EXPAND_SETTINGS_PANEL | 展开设置面板 |
| 7 | COLLAPSE_PANELS | 收起面板 |
| 8 | GET_CLIPBOARD | 获取剪贴板 |
| 9 | SET_CLIPBOARD | 设置剪贴板 |
| 10 | SET_DISPLAY_POWER | 设置显示电源 |
| 11 | ROTATE_DEVICE | 旋转设备 |
| 12 | UHID_CREATE | 创建虚拟 HID |
| 13 | UHID_INPUT | 虚拟 HID 输入 |
| 14 | UHID_DESTROY | 销毁虚拟 HID |
| 15 | OPEN_HARD_KEYBOARD_SETTINGS | 打开键盘设置 |
| 16 | START_APP | 启动应用 |
| 17 | RESET_VIDEO | 重置视频编码 |
| 18 | CAMERA_SET_TORCH | 设置闪光灯 |
| 19 | CAMERA_ZOOM_IN | 变焦放大 |
| 20 | CAMERA_ZOOM_OUT | 变焦缩小 |
| 21 | RESIZE_DISPLAY | 调整显示尺寸 |
| 22 | SCAN_FILE | 扫描文件 |
| 101 | VIDEO_SETTINGS (扩展) | 动态调整视频参数 |
| 102 | FILE_PUSH (扩展) | 文件推送 |

### 设备消息 (Server → Client)

| ID | 消息 | 说明 |
|----|------|------|
| 0 | CLIPBOARD | 剪贴板内容 |
| 1 | ACK_CLIPBOARD | 剪贴板确认 |
| 2 | UHID_OUTPUT | 虚拟 HID 输出 |

详细协议规范见 [docs/PROTOCOL.md](docs/PROTOCOL.md)。

---

## 许可证

[MIT](LICENSE)

---

<div align="center">

**Go-scrcpy** — 用 Go 语言掌控你的 Android 设备

</div>