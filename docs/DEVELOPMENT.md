# Go-scrcpy 开发计划

## 项目概述

Go-scrcpy 是一个可复用的 Go library，实现 scrcpy 协议客户端，通过 ADB 连接 Android 设备进行屏幕镜像、远程控制和文件管理。

- **定位**: Go library + CLI 示例
- **发布**: `github.com/yuan71058/go-scrcpy`
- **协议版本**: scrcpy v4.1
- **Go 版本**: >= 1.22

## 目标

1. 对接 scrcpy-server 全功能（22 种控制消息 + ws-scrcpy 扩展）
2. 支持多设备同时接入控制
3. 作为可被其他 Go 项目引用的 GitHub 库
4. 提供完善的 API 文档和使用示例

---

## 项目结构

```
go-scrcpy/
├── go.mod
├── go.sum
├── README.md
├── DEVELOPMENT.md              # 本文件
├── PROTOCOL.md                 # 二进制协议规范
├── pkg/
│   ├── scrcpy/                 # 核心入口，对外 API
│   │   ├── client.go           # 单设备客户端
│   │   ├── multi.go            # 多设备管理器
│   │   ├── session.go          # 设备会话
│   │   └── options.go          # 启动选项
│   ├── adb/                    # ADB 设备管理
│   │   ├── client.go           # ADB 命令封装
│   │   ├── device.go           # 设备发现与跟踪
│   │   └── forward.go          # 端口转发
│   ├── server/                 # scrcpy-server 生命周期
│   │   ├── launcher.go         # 启动/停止 server
│   │   ├── options.go          # server 参数序列化
│   │   └── version.go          # 版本管理
│   ├── transport/              # 三通道 socket 通信
│   │   ├── connection.go       # 连接管理
│   │   ├── reader.go           # 带缓冲的协议读取器
│   │   └── writer.go           # 协议写入器
│   ├── protocol/               # 协议编解码
│   │   ├── handshake.go        # 握手解析
│   │   ├── control_msg.go      # 控制消息序列化（22+2 种）
│   │   ├── device_msg.go       # 设备消息反序列化
│   │   ├── frame_header.go     # 帧元数据解析
│   │   ├── codec_id.go         # 编解码器 ID 常量
│   │   └── display_info.go     # 显示信息结构
│   ├── video/                  # 视频处理
│   │   ├── decoder.go          # 解码器接口
│   │   ├── nalu.go             # NALU 解析
│   │   ├── h264.go             # H.264 解码
│   │   ├── h265.go             # H.265 解码
│   │   └── renderer.go         # 渲染器接口
│   ├── audio/                  # 音频处理
│   │   ├── decoder.go          # 解码器接口
│   │   ├── opus.go             # Opus 解码
│   │   └── pcm.go              # PCM 播放
│   ├── input/                  # 输入事件
│   │   ├── touch.go            # 触摸事件
│   │   ├── keycode.go          # 按键事件
│   │   ├── keycode_map.go      # Android keycode 映射表
│   │   ├── scroll.go           # 滚动事件
│   │   └── text.go             # 文本注入
│   ├── control/                # 高级控制
│   │   ├── clipboard.go        # 剪贴板同步
│   │   ├── filepush.go         # 文件推送/APK 安装
│   │   └── display.go          # 显示控制（通知栏、旋转等）
│   ├── record/                 # 录制与截图
│   │   ├── recorder.go         # MP4/MKV 录制
│   │   └── screenshot.go       # PNG/JPEG 截图
│   └── types/                  # 公共类型
│       ├── device.go           # DeviceInfo, DisplayInfo
│       ├── video.go            # VideoFrame, VideoCodec
│       └── audio.go            # AudioPacket, AudioCodec
├── cmd/
│   └── go-scrcpy/
│       └── main.go             # CLI 示例入口
└── _examples/
    ├── single/main.go          # 单设备示例
    ├── multi/main.go           # 多设备示例
    ├── record/main.go          # 录制示例
    └── control/main.go         # 远程控制示例
```

---

## 模块详细设计

### 1. ADB 管理 (`pkg/adb/`)

#### Client

```go
type Client struct {
    ExecPath string // adb 可执行文件路径，默认 "adb"
}
```

#### 核心方法

| 方法 | 说明 |
|------|------|
| `ListDevices() ([]Device, error)` | 列举已连接设备 |
| `TrackDevices(ctx) (<-chan DeviceEvent, error)` | 实时跟踪设备上下线 |
| `Forward(serial, remote string) (int, error)` | ADB 端口转发，返回本地端口 |
| `RemoveForward(serial string, port int) error` | 移除端口转发 |
| `Push(serial, local, remote string) error` | 推送文件到设备 |
| `Shell(serial, command string) (string, error)` | 执行 shell 命令 |
| `Pull(serial, remote, local string) error` | 从设备拉取文件 |
| `GetDeviceProperties(serial string) (map[string]string, error)` | 获取设备属性 |
| `GetModel(serial string) string` | 获取设备型号 |
| `GetAndroidVersion(serial string) int` | 获取 Android 版本号 |

#### 多设备关键

- 所有操作通过 `serial` 参数区分设备
- `Forward` 使用 `portfinder` 动态分配本地端口，避免冲突
- 两个设备可同时 forward 到同一个 remote（如 `localabstract:scrcpy`），各自获得不同本地端口

### 2. Server 管理 (`pkg/server/`)

#### Options 结构体

包含 server 启动所需的全部 50+ 参数：

**视频参数:**
- `Video bool` - 启用视频捕获（默认 true）
- `VideoCodec string` - 编解码器: "h264", "h265", "av1", "vp8", "vp9"
- `VideoSource string` - 视频源: "display", "camera"
- `VideoBitRate int` - 码率（默认 8000000）
- `MaxFps float64` - 最大帧率
- `MaxSize int` - 最大分辨率
- `Crop string` - 裁剪: "width:height:x:y"
- `Angle float64` - 旋转角度
- `DisplayID int` - 显示器 ID
- `VideoEncoder string` - 指定编码器名称
- `VideoCodecOptions map[string]string` - 编码器参数

**音频参数:**
- `Audio bool` - 启用音频（默认 true）
- `AudioCodec string` - 编解码器: "opus", "aac", "flac", "raw"
- `AudioSource string` - 音频源: "output", "mic", "playback", "voice-call" 等 10 种
- `AudioBitRate int` - 码率（默认 128000）
- `AudioEncoder string` - 指定编码器
- `AudioCodecOptions map[string]string` - 编码器参数
- `AudioDup bool` - 音频复制

**摄像头参数:**
- `CameraID string` - 摄像头 ID
- `CameraSize string` - 分辨率: "WxH"
- `CameraFacing string` - 方向: "front", "back", "external"
- `CameraZoom float64` - 变焦倍数
- `CameraFPS int` - 帧率
- `CameraHighSpeed bool` - 高速模式
- `CameraTorch bool` - 闪光灯

**控制参数:**
- `Control bool` - 启用控制（默认 true）
- `ClipboardAutosync bool` - 剪贴板自动同步（默认 true）
- `ShowTouches bool` - 显示触摸点
- `StayAwake bool` - 保持屏幕常亮
- `ScreenOffTimeout int` - 屏幕超时（ms）
- `PowerOffOnClose bool` - 关闭时息屏
- `PowerOn bool` - 启动时亮屏（默认 true）
- `KeepActive bool` - 保持设备活跃

**显示参数:**
- `NewDisplay string` - 创建虚拟显示: "WxH[/dpi]"
- `CaptureOrientation string` - 捕获方向: "@0", "@1" 等
- `DisplayIMEPolicy string` - 输入法策略: "local", "fallback", "hide"
- `VDDestroyContent bool` - 虚拟显示销毁内容
- `VDSystemDecorations bool` - 虚拟显示系统装饰
- `FlexDisplay bool` - 弹性显示

**高级参数:**
- `TunnelForward bool` - 使用转发隧道
- `DownsizeOnError bool` - 编码出错时降分辨率（默认 true）
- `Cleanup bool` - 退出时清理（默认 true）
- `IgnoreVideoEncoderConstraints bool` - 忽略编码器限制
- `SendFrameMeta bool` - 发送帧 PTS（默认 true）
- `SCID int` - 会话客户端 ID

**查询参数:**
- `ListEncoders bool` - 列举可用编码器
- `ListDisplays bool` - 列举显示器
- `ListCameras bool` - 列举摄像头
- `ListCameraSizes bool` - 列举摄像头分辨率
- `ListApps bool` - 列举已安装应用

#### Launcher

```go
type Launcher struct {
    ADB *adb.Client
}

func (l *Launcher) Start(ctx context.Context, serial string, opts Options) (*ServerConn, error)
func (l *Launcher) Kill(serial string) error
func (l *Launcher) IsRunning(serial string) (bool, error)
```

#### 启动流程

1. 检查 server 是否已在运行（`ps | grep scrcpy`）
2. 如果未运行，push `scrcpy-server.jar` 到 `/data/local/tmp/`
3. 通过 `adb shell app_process` 启动 server 进程
4. 等待 PID 文件或 socket 就绪
5. 建立 ADB 端口转发到 server 的三个 socket
6. 返回连接信息

### 3. 传输层 (`pkg/transport/`)

#### Connection

```go
type Connection struct {
    VideoConn   net.Conn
    AudioConn   net.Conn
    ControlConn net.Conn
}
```

#### 读写器

```go
type ProtocolReader struct {
    r   *bufio.Reader
    buf []byte
}

func NewProtocolReader(r io.Reader) *ProtocolReader
func (pr *ProtocolReader) ReadFull(n int) ([]byte, error)
func (pr *ProtocolReader) ReadByte() (byte, error)
func (pr *ProtocolReader) ReadUint16() (uint16, error)
func (pr *ProtocolReader) ReadUint32() (uint32, error)
func (pr *ProtocolReader) ReadUint64() (uint64, error)
func (pr *ProtocolReader) ReadInt32() (int32, error)
func (pr *ProtocolReader) ReadInt64() (int64, error)
```

### 4. 协议层 (`pkg/protocol/`)

#### 握手

```go
type Handshake struct {
    DeviceName string       // 64B, null-padded UTF-8
    Displays   []DisplayInfo
    Encoders   []string
    ClientID   int32
}

type DisplayInfo struct {
    DisplayID  int32
    Width      int32
    Height     int32
    Rotation   int32
    LayerStack int32
    Flags      int32
}

// 每个 display 还附带:
//   ConnectionCount int32     - 当前连接数
//   ScreenInfo      []byte    - 25B (contentRect + videoSize + rotation)
//   VideoSettings   []byte    - 35B+ (bitrate, fps, bounds, crop, etc.)
```

#### 控制消息（24 种）

| ID | 名称 | 载荷大小 | 说明 |
|----|------|----------|------|
| 0 | INJECT_KEYCODE | 14B | [action:1][keycode:4][repeat:4][meta:4] |
| 1 | INJECT_TEXT | 5+len | [len:4][utf8 text] |
| 2 | INJECT_TOUCH_EVENT | 32B | [action:1][pointerId:8][x:4][y:4][screenW:2][screenH:2][pressure:2][actionBtn:4][buttons:4] |
| 3 | INJECT_SCROLL_EVENT | 21B | [x:4][y:4][screenW:2][screenH:2][hscroll:2][vscroll:2][buttons:4] |
| 4 | BACK_OR_SCREEN_ON | 2B | [action:1] |
| 5 | EXPAND_NOTIFICATION_PANEL | 1B | 无载荷 |
| 6 | EXPAND_SETTINGS_PANEL | 1B | 无载荷 |
| 7 | COLLAPSE_PANELS | 1B | 无载荷 |
| 8 | GET_CLIPBOARD | 2B | [copyKey:1] |
| 9 | SET_CLIPBOARD | 10+len | [sequence:8][paste:1][len:4][utf8 text] |
| 10 | SET_DISPLAY_POWER | 2B | [on:1] |
| 11 | ROTATE_DEVICE | 1B | 无载荷 |
| 12 | UHID_CREATE | 9+ | [id:2][vendorId:2][productId:2][nameLen:1][name][descSize:2][desc] |
| 13 | UHID_INPUT | 5+ | [id:2][size:2][data] |
| 14 | UHID_DESTROY | 3B | [id:2] |
| 15 | OPEN_HARD_KEYBOARD_SETTINGS | 1B | 无载荷 |
| 16 | START_APP | 2+len | [nameLen:1][packageName] |
| 17 | RESET_VIDEO | 1B | 无载荷 |
| 18 | CAMERA_SET_TORCH | 2B | [on:1] |
| 19 | CAMERA_ZOOM_IN | 1B | 无载荷 |
| 20 | CAMERA_ZOOM_OUT | 1B | 无载荷 |
| 21 | RESIZE_DISPLAY | 5B | [width:2][height:2] |
| 22 | SCAN_FILE | 5+len | [pathLen:4][path] |
| 101 | VIDEO_SETTINGS | 35+ | VideoSettings 动态调整 (ws-scrcpy 扩展) |
| 102 | FILE_PUSH | 变长 | 文件推送 NEW/START/APPEND/FINISH (ws-scrcpy 扩展) |

#### 设备消息（3 种）

| ID | 名称 | 载荷格式 |
|----|------|----------|
| 0 | CLIPBOARD | [length:4][utf8 text] |
| 1 | ACK_CLIPBOARD | [sequence:8] |
| 2 | UHID_OUTPUT | [id:2][size:2][data] |

#### 帧元数据（per-packet header, 12B）

```
[8B] PTS + flags:
  bit 62 = CONFIG packet (SPS/PPS)
  bit 61 = keyframe
  其余   = PTS 微秒时间戳
[4B] packet size
```

#### 编解码器 ID（4 字节 ASCII）

**视频:**
- `h264` (0x68323634)
- `h265` (0x68323635)
- `\0av1` (0x00617631)
- `\0vp8` (0x00767038)
- `\0vp9` (0x00767039)

**音频:**
- `opus` (0x6F707573)
- `\0aac` (0x00616163)
- `flac` (0x666C6163)
- `\0raw` (0x00726177)

### 5. 多设备管理 (`pkg/scrcpy/`)

#### Client（单设备）

```go
type Client struct {
    session  *DeviceSession
    opts     server.Options
    serial   string
    adb      *adb.Client
    launcher *server.Launcher
}

func New(serial string, opts server.Options) *Client
func (c *Client) Start(ctx context.Context) error
func (c *Client) Close() error
func (c *Client) VideoStream() <-chan types.VideoFrame
func (c *Client) AudioStream() <-chan types.AudioPacket
func (c *Client) SendControl(msg []byte) error
func (c *Client) DeviceInfo() types.DeviceInfo
func (c *Client) Handshake() *protocol.Handshake
```

#### MultiClient（多设备）

```go
type MultiClient struct {
    clients map[string]*Client // keyed by serial
    mu      sync.RWMutex
    adb     *adb.Client
}

func NewMulti(adbClient *adb.Client) *MultiClient
func (m *MultiClient) Add(serial string, opts server.Options) (*Client, error)
func (m *MultiClient) Remove(serial string) error
func (m *MultiClient) Get(serial string) (*Client, bool)
func (m *MultiClient) List() []*Client
func (m *MultiClient) Count() int
func (m *MultiClient) Broadcast(msg []byte) // 向所有设备发送
func (m *MultiClient) Close() error         // 关闭所有连接
```

#### DeviceWatcher（自动发现）

```go
type DeviceWatcher struct {
    adb      *adb.Client
    onAdd    func(serial string)
    onRemove func(serial string)
}

func NewDeviceWatcher(adbClient *adb.Client) *DeviceWatcher
func (w *DeviceWatcher) OnAdd(fn func(serial string))
func (w *DeviceWatcher) OnRemove(fn func(serial string))
func (w *DeviceWatcher) Start(ctx context.Context) error
```

#### 多设备并发架构

```
MultiClient
├── Client[A] → Session → VideoConn + AudioConn + ControlConn → Device A
├── Client[B] → Session → VideoConn + AudioConn + ControlConn → Device B
└── Client[C] → Session → VideoConn + AudioConn + ControlConn → Device C
         ↓
    ADB forward 每设备独立端口（portfinder 自动分配）
    goroutine per device（视频/音频/控制独立 goroutine）
    context 控制生命周期
```

### 6. 视频处理 (`pkg/video/`)

#### 解码器接口

```go
type Decoder interface {
    Push(data []byte) error
    ReadFrame() (*types.VideoFrame, error)
    Close() error
}

func NewH264Decoder() Decoder
func NewH265Decoder() Decoder
```

#### NALU 解析

```go
type NALUType int
const (
    NALUTypeSPS NALUType = 7
    NALUTypePPS NALUType = 8
    NALUTypeIDR NALUType = 5
    NALUTypeNonIDR NALUType = 1
)

type NALU struct {
    Type    NALUType
    Data    []byte
    Temporal int
}

type NALUReader struct { ... }
func (r *NALUReader) Next() (NALU, error)
func ParseSPS(data []byte) (*SPSInfo, error) // 提取宽高、profile、level
```

#### 渲染器接口

```go
type Renderer interface {
    Render(frame *types.VideoFrame) error
    Close() error
}
```

### 7. 音频处理 (`pkg/audio/`)

```go
type Decoder interface {
    Push(data []byte) error
    ReadPCM() ([]byte, error)
    Close() error
}

type Player interface {
    Play(pcm []byte) error
    Close() error
}

func NewOpusDecoder() Decoder
func NewAACDecoder() Decoder
func NewPCMPlayer() Player
```

### 8. 输入事件 (`pkg/input/`)

```go
// 每个函数返回可直接通过 control channel 发送的 []byte
func TouchDown(pointerID uint64, x, y int32, screenW, screenH uint16) []byte
func TouchMove(pointerID uint64, x, y int32, screenW, screenH uint16) []byte
func TouchUp(pointerID uint64, x, y int32, screenW, screenH uint16) []byte

func KeyDown(keycode int32, meta int32) []byte
func KeyUp(keycode int32, meta int32) []byte
func KeyPress(keycode int32, meta int32) []byte // DOWN + UP

func Text(text string) []byte
func Scroll(x, y int32, screenW, screenH uint16, hscroll, vscroll int16) []byte

func BackOrScreenOn() []byte
func ExpandNotificationPanel() []byte
func ExpandSettingsPanel() []byte
func CollapsePanels() []byte
func RotateDevice() []byte
func SetDisplayPower(on bool) []byte
func ResetVideo() []byte
func StartApp(packageName string) []byte
func ScanFile(path string) []byte

// 摄像头控制
func CameraSetTorch(on bool) []byte
func CameraZoomIn() []byte
func CameraZoomOut() []byte
func ResizeDisplay(width, height uint16) []byte
```

#### Android Keycode 映射表

完整的 `KEYCODE_*` 常量映射，从 `KeyEvent.ts` 翻译为 Go 常量。

### 9. 控制功能 (`pkg/control/`)

#### 剪贴板

```go
type Clipboard struct {
    conn     *transport.Connection
    sequence uint64
    mu       sync.Mutex
    onChange  func(text string)
}

func NewClipboard(conn *transport.Connection) *Clipboard
func (c *Clipboard) Get(ctx context.Context) (string, error)
func (c *Clipboard) Set(text string, paste bool) error
func (c *Clipboard) OnChange(fn func(text string))
```

#### 文件推送

```go
type FilePusher struct {
    conn *transport.Connection
}

type PushProgress struct {
    ID       int16
    Code     byte
    BytesSent int64
    Total    int64
}

func NewFilePusher(conn *transport.Connection) *FilePusher
func (fp *FilePusher) PushFile(ctx context.Context, filename string, data io.Reader, progress func(PushProgress)) error
func (fp *FilePusher) InstallAPK(ctx context.Context, apkPath string) error
```

#### 文件推送协议

```
NEW     → Server 分配 push ID（返回 DeviceMessage type=101）
START   → 发送 filename + fileSize
APPEND  → 分块发送文件内容
FINISH  → 信号完成
```

### 10. 录制与截图 (`pkg/record/`)

```go
type Recorder struct {
    w         io.Writer
    muxer     *NareixMuxer
    videoTrack int
    audioTrack int
}

func NewRecorder(w io.Writer, format string) *Recorder // "mp4" or "mkv"
func (r *Recorder) WriteVideo(data []byte, pts int64, keyframe bool) error
func (r *Recorder) WriteAudio(data []byte, pts int64) error
func (r *Recorder) Close() error

func Screenshot(frame *types.VideoFrame, format string) ([]byte, error) // "png" or "jpeg"
```

---

## 依赖库

| 用途 | 库 | 说明 |
|------|-----|------|
| ADB | `os/exec` 调用 adb | 轻量，无额外依赖 |
| 视频解码 | `github.com/nareix/joy5` | 纯 Go，支持 H.264 解码 |
| 音频解码 | `github.com/nareix/joy5` | Opus/AAC 解码 |
| MP4 封装 | `github.com/nareix/joy5` | 录制封装 |
| CLI | `github.com/spf13/cobra` | 仅 CLI 示例使用 |
| UUID | `github.com/google/uuid` | session ID 生成 |

---

## 开发里程碑

| 阶段 | 内容 | 验证标准 |
|------|------|----------|
| **M1** | ADB + Server + Handshake | `adb.go` 列举设备、启动 server、完成握手获取设备名和显示器信息 |
| **M2** | 单设备视频 | 能解码并输出 H.264 视频帧 |
| **M3** | 单设备音频 | 能解码并输出 Opus/PCM 音频 |
| **M4** | 单设备输入 | 能发送触控/按键/滚动事件 |
| **M5** | 多设备并发 | 3 台设备同时接入，独立控制，互不影响 |
| **M6** | 剪贴板 + 文件推送 | 双向剪贴板同步、APK 安装 |
| **M7** | 录制 + 截图 | MP4 录制、PNG 截图 |
| **M8** | 摄像头 + UHID | 前后摄切换、torch/zoom、虚拟 HID |
| **M9** | Library 发布 | godoc 完善、example 完整、go get 可用 |

---

## API 设计原则

1. **接口优于实现**: `Decoder`, `AudioDecoder`, `Renderer` 定义为接口，用户可替换
2. **Option 模式**: 所有配置通过 `server.Options` struct，零值即合理默认
3. **Channel 驱动**: 视频帧、音频包通过 Go channel 流式传递
4. **Context 贯穿**: 所有阻塞操作接受 ctx，支持超时和取消
5. **错误不隐藏**: 所有错误返回 error，不 panic
6. **无 CGo 必选**: 核心库纯 Go，解码器可选 CGo 实现
7. **并发安全**: MultiClient 操作加锁，per-device session 无锁

---

## 多设备使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
    "github.com/yuan71058/go-scrcpy/pkg/server"
)

func main() {
    adbClient := &adb.Client{ExecPath: "adb"}

    // 多设备管理器
    mc := scrcpy.NewMulti(adbClient)
    defer mc.Close()

    // 自动发现设备并接入
    watcher := scrcpy.NewDeviceWatcher(adbClient)
    watcher.OnAdd(func(serial string) {
        fmt.Printf("设备上线: %s\n", serial)
        c, err := mc.Add(serial, server.Options{
            VideoBitRate: 4000000,
            Audio:        true,
            MaxFps:       30,
        })
        if err != nil {
            log.Printf("接入设备 %s 失败: %v", serial, err)
            return
        }

        // 每个设备独立 goroutine 处理视频
        go func() {
            for frame := range c.VideoStream() {
                _ = frame // 处理视频帧
            }
        }()

        // 每个设备独立 goroutine 处理音频
        go func() {
            for pkt := range c.AudioStream() {
                _ = pkt // 处理音频包
            }
        }()
    })

    watcher.OnRemove(func(serial string) {
        fmt.Printf("设备离线: %s\n", serial)
        mc.Remove(serial)
    })

    // 启动设备跟踪
    if err := watcher.Start(context.Background()); err != nil {
        log.Fatal(err)
    }

    // 向所有设备广播通知栏展开
    mc.Broadcast(input.ExpandNotificationPanel())
}
```

---

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 编码器不兼容 | 视频无法解码 | 回退到默认编码器，ListEncoders 查询可用列表 |
| Android 版本差异 | AudioCapture 需 Android 11+，UHID 需 Android 9+ | 检查版本号，功能降级 |
| H.265/AV1 解码复杂 | 纯 Go 解码困难 | 可选 CGo + FFmpeg，或仅支持 H.264 |
| GC 停顿影响帧率 | 视频卡顿 | 使用 pool 复用 buffer，减少分配 |
| ADB 连接不稳定 | 设备断连 | 自动重连 + goroutine leak 检测 |
| 多设备端口冲突 | 连接失败 | portfinder 动态分配，检查已有 forward |
