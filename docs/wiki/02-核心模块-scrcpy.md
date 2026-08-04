# Go-scrcpy Code Wiki — 核心 API（pkg/scrcpy）

`pkg/scrcpy` 是 Go-scrcpy 库的核心入口包，对外暴露单设备客户端（`Client`）、多设备管理器（`MultiClient`）、设备会话（`DeviceSession`）和配置选项（`Options`）。

## 包结构

```
pkg/scrcpy/
├── log.go        # 日志级别控制
├── options.go    # Options 配置 + 链式调用
├── client.go     # Client 单设备客户端
├── session.go    # DeviceSession 设备会话
└── multi.go      # MultiClient 多设备管理器 + 辅助函数
```

## 类型定义

### Options — 配置选项

```go
type Options struct {
    ADBPath  string        // ADB 可执行文件路径，默认 "adb"
    LocalJAR string        // 本地 scrcpy-server.jar 路径（为空则使用设备上已有的）
    Server   server.Options // Server 启动参数
}
```

**DefaultOptions()** 返回包含合理默认值的配置：

```go
func DefaultOptions() Options {
    return Options{
        ADBPath: "adb",
        Server: server.Options{
            Video:            true,    // 启用视频
            Audio:            true,    // 启用音频
            Control:          true,    // 启用控制
            VideoBitRate:     8000000, // 8Mbps
            AudioBitRate:     128000,  // 128kbps
            ClipboardAutosync: true,   // 剪贴板自动同步
            DownsizeOnError:  true,    // 编码出错时降分辨率
            Cleanup:          true,    // 退出时清理
            PowerOn:          true,    // 启动时亮屏
            SendFrameMeta:    true,    // 发送帧 PTS
            SCID:             -1,      // 不设置会话客户端 ID
            ScreenOffTimeout: -1,      // 不设置屏幕超时
        },
    }
}
```

**链式调用方法：**

| 方法 | 说明 |
|------|------|
| `WithVideoCodec(codec)` | 设置视频编解码器 |
| `WithAudioCodec(codec)` | 设置音频编解码器 |
| `WithMaxFps(fps)` | 设置最大帧率 |
| `WithMaxSize(size)` | 设置最大分辨率 |
| `WithBitrate(bitrate)` | 设置视频码率 |
| `WithAudioEnabled(enabled)` | 启用/禁用音频 |
| `WithControlEnabled(enabled)` | 启用/禁用控制 |

### Client — 单设备客户端

```go
type Client struct {
    session    *DeviceSession  // 设备会话
    adb        *adb.Client     // ADB 客户端
    opts       Options         // 启动选项
    listenPort int             // PC 监听端口
    closed     bool
}
```

**构造函数：**

```go
func New(serial string, opts Options, listenPort int) *Client
```

- `serial`: 设备序列号
- `opts`: 启动选项
- `listenPort`: PC 监听端口（0 = 自动分配）

**生命周期方法：**

| 方法 | 说明 |
|------|------|
| `Start(ctx) error` | 启动客户端，建立连接并开始接收流 |
| `Close() error` | 关闭客户端 |

**流获取方法：**

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `VideoStream()` | `<-chan *video.DecodedFrame` | 视频帧通道 |
| `AudioStream()` | `<-chan *audio.AudioData` | 音频数据通道 |

**信息查询方法：**

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `Serial()` | `string` | 设备序列号 |
| `DeviceInfo()` | `*types.DeviceInfo` | 设备信息 |
| `Handshake()` | `*protocol.Handshake` | 握手数据 |

**输入方法（均为便捷包装，内部调用 `SendControl`）：**

| 分类 | 方法 |
|------|------|
| 触摸 | `TouchDown`, `TouchMove`, `TouchUp` |
| 鼠标 | `MouseDown`, `MouseMove`, `MouseUp` |
| 滚动 | `Scroll` |
| 按键 | `KeyDown`, `KeyUp`, `KeyPress`, `Text` |
| 快捷 | `Home`, `Back`, `Power`, `VolumeUp`, `VolumeDown`, `Menu`, `Enter`, `Delete`, `Tab`, `Escape`, `Space` |
| 面板 | `ExpandNotificationPanel`, `ExpandSettingsPanel`, `CollapsePanels`, `RotateDevice` |
| 显示 | `SetDisplayPower`, `PowerOn`, `PowerOff`, `ResizeDisplay` |
| 剪贴板 | `SetClipboard`, `GetClipboard` |
| 应用 | `StartApp`, `ScanFile` |
| 摄像头 | `CameraSetTorch`, `CameraZoomIn`, `CameraZoomOut` |
| UHID | `UHIDCreate`, `UHIDInput`, `UHIDDestroy` |
| 其他 | `ResetVideo` |

### DeviceSession — 设备会话

```go
type DeviceSession struct {
    serial     string                // 设备序列号
    listener   *transport.Listener   // reverse tunnel 监听器
    conn       *transport.Connection  // 三通道连接
    handshake  *protocol.Handshake    // 握手数据
    videoDec   *video.H264Decoder     // 视频解码器
    audioDec   *audio.OpusDecoder     // 音频解码器
    clipboard  *control.Clipboard     // 剪贴板管理器
    filePusher *control.FilePusher    // 文件推送器
    deviceInfo *types.DeviceInfo      // 设备信息
    videoChan  chan *video.DecodedFrame // 视频帧通道
    audioChan  chan *audio.AudioData    // 音频数据通道
    controlChan chan []byte             // 控制消息通道
    mu         sync.Mutex
    closed     bool
}
```

**关键方法：**

| 方法 | 说明 |
|------|------|
| `Connect(ctx, adb, opts, jar, port)` | 建立连接（创建监听器 → 启动server → 接受连接 → 握手 → 初始化） |
| `Start(ctx)` | 启动会话（开启3个 goroutine 分别读取视频/音频/控制） |
| `Close()` | 关闭会话（按顺序关闭解码器、连接、监听器、通道） |

**启动流程：**

```
Client.Start(ctx)
  └── DeviceSession.Connect(ctx, adb, opts, jar, port)
        ├── 1. transport.NewListener(port)          // 创建 TCP 监听器
        ├── 2. server.NewLauncher(adb).Start(...)    // 推送 JAR、adb reverse、启动进程
        ├── 3. listener.Accept(video, audio, control) // 等待 Android 连接 1~3 个 socket
        ├── 4. NewConnectionFromListener(listener)    // 创建三通道连接
        ├── 5. protocol.ReadHandshake(videoReader)    // 读取设备名、分辨率
        ├── 6. video.NewH264Decoder(60)               // 创建视频解码器
        ├── 7. audio.NewOpusDecoder(100)              // 创建音频解码器（可选）
        ├── 8. control.NewClipboard(session)           // 创建剪贴板管理器
        ├── 9. control.NewFilePusher(session)          // 创建文件推送器
        └── 10. adb.GetDeviceProperties()              // 获取设备属性
  └── DeviceSession.Start(ctx)
        ├── go readVideoLoop(ctx)      // 读取视频帧循环
        ├── go readAudioLoop(ctx)      // 读取音频帧循环（可选）
        └── go readControlLoop(ctx)    // 读取控制消息循环
```

**读取循环：**

| 循环 | 数据源 | 处理流程 |
|------|--------|----------|
| `readVideoLoop` | 视频 socket | 读帧头(12B) → 读帧数据 → `videoDec.Push()` → `videoChan <- frame` |
| `readAudioLoop` | 音频 socket | 读帧头(12B) → 读帧数据 → `audioDec.Push()` → `audioChan <- pkt` |
| `readControlLoop` | 控制 socket | 读设备消息 → `dispatchDeviceMessage()` |

### MultiClient — 多设备管理器

```go
type MultiClient struct {
    clients map[string]*Client  // keyed by serial
    adb     *adb.Client
    mu      sync.RWMutex
    closed  bool
}
```

**方法：**

| 方法 | 说明 |
|------|------|
| `NewMulti(adbClient)` | 创建多设备管理器 |
| `Add(serial, opts)` | 添加设备（自动分配端口并启动） |
| `Remove(serial)` | 移除设备 |
| `Get(serial)` | 获取指定设备客户端 |
| `List()` | 列出所有设备 |
| `Count()` | 获取设备数量 |
| `Broadcast(msg)` | 向所有设备广播控制消息 |
| `ForEach(fn)` | 遍历所有设备执行操作 |
| `VideoStreamAll()` | 合并所有设备视频帧到一个通道 |
| `Close()` | 关闭所有设备连接 |

### 辅助函数

| 函数 | 说明 |
|------|------|
| `GetDevices(adbClient)` | 获取已连接的设备列表 |
| `WatchDevices(adbClient, onAdd, onRemove)` | 监听设备上下线（便捷包装器） |

### VideoFrameWithSerial

```go
type VideoFrameWithSerial struct {
    Serial string             // 设备序列号
    Frame  *video.DecodedFrame // 视频帧
}
```

## 日志系统

每个模块有独立的日志级别控制：

```go
const (
    LogLevelNone  = iota  // 无日志
    LogLevelError         // 仅错误
    LogLevelInfo          // 信息+错误
    LogLevelDebug         // 全部
)

func SetLogLevel(level int)
```

内部日志函数：
- `logDebug(format, args...)` — 调试级别
- `logInfo(format, args...)` — 信息级别
- `logError(format, args...)` — 错误级别