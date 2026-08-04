# Go-scrcpy API 文档

## 目录

- [包概览](#包概览)
- [pkg/scrcpy](#pkgscrcpy)
- [pkg/adb](#pkgadb)
- [pkg/server](#pkgserver)
- [pkg/transport](#pkgtransport)
- [pkg/protocol](#pkgprotocol)
- [pkg/video](#pkgvideo)
- [pkg/audio](#pkgaudio)
- [pkg/input](#pkginput)
- [pkg/control](#pkgcontrol)
- [pkg/record](#pkgrecord)
- [pkg/types](#pkgtypes)

---

## 包概览

Go-scrcpy 是一个 Go 语言实现的 scrcpy 客户端库，用于通过 scrcpy 协议连接和控制 Android 设备。

### 依赖关系

```
pkg/scrcpy (核心入口)
├── pkg/adb (ADB 设备管理)
├── pkg/server (Server 启动器)
├── pkg/transport (Socket 连接)
├── pkg/protocol (协议编解码)
├── pkg/video (视频解码)
├── pkg/audio (音频解码)
├── pkg/input (输入事件)
├── pkg/control (控制功能)
├── pkg/record (录制截图)
└── pkg/types (公共类型)
```

---

## pkg/scrcpy

核心 API 包，提供单设备和多设备连接控制功能。

### 常量

```go
// 日志级别
const (
    LogLevelNone  = iota // 无日志
    LogLevelError        // 错误日志
    LogLevelInfo         // 信息日志
    LogLevelDebug        // 调试日志
)
```

### 函数

#### SetLogLevel

```go
func SetLogLevel(level int)
```

设置 scrcpy 模块的日志级别。

**参数：**
- `level`: 日志级别常量

**示例：**
```go
scrcpy.SetLogLevel(scrcpy.LogLevelDebug)
```

---

### Options 结构体

```go
type Options struct {
    ADBPath  string        // ADB 可执行文件路径，默认 "adb"
    LocalJAR string        // 本地 scrcpy-server.jar 路径
    Server   server.Options // Server 启动参数
}
```

#### DefaultOptions

```go
func DefaultOptions() Options
```

返回默认配置选项。

**返回值：** 包含合理默认值的 Options 结构体

#### WithVideoCodec

```go
func (o Options) WithVideoCodec(codec string) Options
```

设置视频编解码器（链式调用）。

**参数：**
- `codec`: 编解码器名称 ("h264", "h265", "av1", "vp8", "vp9")

**返回值：** 新的 Options 副本

#### WithAudioCodec

```go
func (o Options) WithAudioCodec(codec string) Options
```

设置音频编解码器（链式调用）。

**参数：**
- `codec`: 编解码器名称 ("opus", "aac", "flac", "raw")

**返回值：** 新的 Options 副本

#### WithMaxFps

```go
func (o Options) WithMaxFps(fps float64) Options
```

设置最大帧率（链式调用）。

**参数：**
- `fps`: 最大帧率，0 表示不限制

**返回值：** 新的 Options 副本

#### WithMaxSize

```go
func (o Options) WithMaxSize(size int) Options
```

设置最大分辨率（链式调用）。

**参数：**
- `size`: 最大分辨率，0 表示不限制

**返回值：** 新的 Options 副本

#### WithBitrate

```go
func (o Options) WithBitrate(bitrate int) Options
```

设置视频码率（链式调用）。

**参数：**
- `bitrate`: 码率（bps）

**返回值：** 新的 Options 副本

#### WithAudioEnabled

```go
func (o Options) WithAudioEnabled(enabled bool) Options
```

启用/禁用音频（链式调用）。

**参数：**
- `enabled`: 是否启用音频

**返回值：** 新的 Options 副本

#### WithControlEnabled

```go
func (o Options) WithControlEnabled(enabled bool) Options
```

启用/禁用控制（链式调用）。

**参数：**
- `enabled`: 是否启用控制

**返回值：** 新的 Options 副本

---

### Client 结构体

单设备客户端，封装与单个 Android 设备的 scrcpy 连接。

#### New

```go
func New(serial string, opts Options, listenPort int) *Client
```

创建新的单设备客户端。

**参数：**
- `serial`: 设备序列号
- `opts`: 启动选项
- `listenPort`: PC 监听端口 (0 = 自动分配)

**返回值：** Client 指针

**示例：**
```go
opts := scrcpy.DefaultOptions()
client := scrcpy.New("emulator-5554", opts, 27183)
```

#### Start

```go
func (c *Client) Start(ctx context.Context) error
```

启动客户端，建立连接并开始接收视频流。

**参数：**
- `ctx`: 上下文，用于控制生命周期

**返回值：** 错误信息

#### VideoStream

```go
func (c *Client) VideoStream() <-chan *video.DecodedFrame
```

返回视频帧通道。

**返回值：** 只读通道，接收视频帧

#### AudioStream

```go
func (c *Client) AudioStream() <-chan *audio.AudioData
```

返回音频数据通道。

**返回值：** 只读通道，接收音频数据

#### SendControl

```go
func (c *Client) SendControl(msg []byte) error
```

发送控制消息。

**参数：**
- `msg`: 控制消息字节切片

**返回值：** 错误信息

#### Serial

```go
func (c *Client) Serial() string
```

获取设备序列号。

**返回值：** 设备序列号字符串

#### DeviceInfo

```go
func (c *Client) DeviceInfo() *types.DeviceInfo
```

获取设备信息。

**返回值：** 设备信息指针

#### Handshake

```go
func (c *Client) Handshake() interface{}
```

获取握手数据（返回 `*protocol.Handshake` 需类型断言）。

**返回值：** 握手数据接口，需类型断言为 `*protocol.Handshake`

**示例：**
```go
hs := client.Handshake().(*protocol.Handshake)
fmt.Printf("设备: %s, 尺寸: %dx%d\n", hs.GetDeviceName(), hs.GetDisplayWidth(), hs.GetDisplayHeight())
```

#### Close

```go
func (c *Client) Close() error
```

关闭客户端。

**返回值：** 错误信息

---

### Client 输入方法

#### 触摸输入

```go
// TouchDown 触摸按下
func (c *Client) TouchDown(pointerID uint64, x, y int32, screenW, screenH uint16) error

// TouchMove 触摸移动
func (c *Client) TouchMove(pointerID uint64, x, y int32, screenW, screenH uint16) error

// TouchUp 触摸抬起
func (c *Client) TouchUp(pointerID uint64, x, y int32, screenW, screenH uint16) error
```

**参数：**
- `pointerID`: 指针 ID (0xFFFFFFFFFFFFFFFF=鼠标, 0xFFFFFFFFFFFFFFFE=通用手指)
- `x, y`: 触摸坐标
- `screenW, screenH`: 屏幕尺寸

#### 鼠标输入

```go
// MouseDown 鼠标按下
func (c *Client) MouseDown(x, y int32, screenW, screenH uint16, button int32) error

// MouseMove 鼠标移动
func (c *Client) MouseMove(x, y int32, screenW, screenH uint16, buttons int32) error

// MouseUp 鼠标抬起
func (c *Client) MouseUp(x, y int32, screenW, screenH uint16, button int32) error
```

**参数：**
- `button`: 鼠标按钮 (1=左键, 2=右键, 4=中键)
- `buttons`: 按钮状态位掩码

#### 滚动

```go
func (c *Client) Scroll(x, y int32, screenW, screenH uint16, hscroll, vscroll int16) error
```

**参数：**
- `hscroll`: 水平滚动量 (正数向右, 负数向左)
- `vscroll`: 垂直滚动量 (正数向下, 负数向上)

#### 按键输入

```go
// KeyDown 按键按下
func (c *Client) KeyDown(keycode int32, meta int32) error

// KeyUp 按键抬起
func (c *Client) KeyUp(keycode int32, meta int32) error

// KeyPress 完整按键事件 (按下 + 抬起)
func (c *Client) KeyPress(keycode int32, meta int32) error

// Text 文本注入
func (c *Client) Text(text string) error
```

**参数：**
- `keycode`: Android KEYCODE_* 常量
- `meta`: 修饰键状态 (MetaShiftLeft | MetaCtrlLeft 等)

#### 快捷按键

```go
func (c *Client) Home() error      // HOME 键
func (c *Client) Back() error      // 返回键
func (c *Client) Power() error     // 电源键
func (c *Client) VolumeUp() error  // 音量加
func (c *Client) VolumeDown() error // 音量减
func (c *Client) Menu() error      // 菜单键
func (c *Client) Enter() error     // 回车键
func (c *Client) Delete() error    // 删除键
func (c *Client) Tab() error       // Tab 键
func (c *Client) Escape() error    // ESC 键
func (c *Client) Space() error     // 空格键
```

---

### Client 面板控制

```go
// ExpandNotificationPanel 展开通知栏
func (c *Client) ExpandNotificationPanel() error

// ExpandSettingsPanel 展开设置面板
func (c *Client) ExpandSettingsPanel() error

// CollapsePanels 收起面板
func (c *Client) CollapsePanels() error

// RotateDevice 旋转设备
func (c *Client) RotateDevice() error
```

---

### Client 显示控制

```go
// SetDisplayPower 设置屏幕电源
func (c *Client) SetDisplayPower(on bool) error

// PowerOn 亮屏
func (c *Client) PowerOn() error

// PowerOff 息屏
func (c *Client) PowerOff() error

// ResizeDisplay 调整显示器大小
func (c *Client) ResizeDisplay(width, height uint16) error
```

---

### Client 剪贴板

```go
// SetClipboard 设置剪贴板
func (c *Client) SetClipboard(text string, paste bool, sequence uint64) error

// GetClipboard 获取剪贴板
func (c *Client) GetClipboard() error
```

**参数：**
- `text`: 剪贴板内容
- `paste`: 是否自动粘贴
- `sequence`: 序列号 (单调递增，用于 ACK)

---

### Client 应用控制

```go
// StartApp 启动应用
func (c *Client) StartApp(packageName string) error

// ScanFile 扫描文件
func (c *Client) ScanFile(path string) error
```

---

### Client 摄像头控制

```go
// CameraSetTorch 设置相机闪光灯
func (c *Client) CameraSetTorch(on bool) error

// CameraZoomIn 相机放大
func (c *Client) CameraZoomIn() error

// CameraZoomOut 相机缩小
func (c *Client) CameraZoomOut() error
```

---

### Client UHID 设备

```go
// UHIDCreate 创建 UHID 设备
func (c *Client) UHIDCreate(id uint16, vendorID, productID uint16, name string, reportDesc []byte) error

// UHIDInput UHID 输入
func (c *Client) UHIDInput(id uint16, data []byte) error

// UHIDDestroy 销毁 UHID 设备
func (c *Client) UHIDDestroy(id uint16) error
```

**参数：**
- `id`: 设备 ID
- `vendorID`: 厂商 ID
- `productID`: 产品 ID
- `name`: 设备名称
- `reportDesc`: HID 报告描述符

---

### Client 其他

```go
// ResetVideo 重置视频
func (c *Client) ResetVideo() error
```

---

### DeviceSession 结构体

设备会话，管理单个设备的连接、流和控制。

#### NewSession

```go
func NewSession(serial string) *DeviceSession
```

创建设备会话。

**参数：**
- `serial`: 设备序列号

**返回值：** DeviceSession 指针

#### Connect

```go
func (s *DeviceSession) Connect(ctx context.Context, adbClient *adb.Client, opts server.Options, localJAR string, listenPort int) error
```

建立与设备的连接 (使用 reverse tunnel 模式)。

**参数：**
- `ctx`: 上下文
- `adbClient`: ADB 客户端
- `opts`: server 启动参数
- `localJAR`: 本地 scrcpy-server.jar 路径
- `listenPort`: PC 监听端口

**返回值：** 错误信息

#### Start

```go
func (s *DeviceSession) Start(ctx context.Context) error
```

启动会话，开始读取视频帧、音频数据和设备消息。

**参数：**
- `ctx`: 上下文

**返回值：** 错误信息

#### VideoStream

```go
func (s *DeviceSession) VideoStream() <-chan *video.DecodedFrame
```

返回视频帧通道。

**返回值：** 只读通道

#### AudioStream

```go
func (s *DeviceSession) AudioStream() <-chan *audio.AudioData
```

返回音频数据通道。

**返回值：** 只读通道

#### SendControl

```go
func (s *DeviceSession) SendControl(msg []byte) error
```

发送控制消息。

**参数：**
- `msg`: 控制消息

**返回值：** 错误信息

#### GetDeviceInfo

```go
func (s *DeviceSession) GetDeviceInfo() *types.DeviceInfo
```

获取设备信息。

**返回值：** 设备信息指针

#### GetHandshake

```go
func (s *DeviceSession) GetHandshake() *protocol.Handshake
```

获取握手数据。

**返回值：** 握手数据指针

#### GetSerial

```go
func (s *DeviceSession) GetSerial() string
```

获取设备序列号。

**返回值：** 序列号字符串

#### GetClipboard

```go
func (s *DeviceSession) GetClipboard() *control.Clipboard
```

获取剪贴板管理器。

**返回值：** 剪贴板管理器指针

#### GetFilePusher

```go
func (s *DeviceSession) GetFilePusher() *control.FilePusher
```

获取文件推送器。

**返回值：** 文件推送器指针

#### Close

```go
func (s *DeviceSession) Close() error
```

关闭会话。

**返回值：** 错误信息

#### IsClosed

```go
func (s *DeviceSession) IsClosed() bool
```

检查会话是否已关闭。

**返回值：** 是否已关闭

---

### MultiClient 结构体

多设备管理器，支持多个设备同时接入控制。

#### NewMulti

```go
func NewMulti(adbClient *adb.Client) *MultiClient
```

创建多设备管理器。

**参数：**
- `adbClient`: ADB 客户端

**返回值：** MultiClient 指针

#### Add

```go
func (m *MultiClient) Add(serial string, opts Options) (*Client, error)
```

添加设备。自动查找可用端口并启动连接。

**参数：**
- `serial`: 设备序列号
- `opts`: 启动选项

**返回值：** 客户端和错误信息

#### Remove

```go
func (m *MultiClient) Remove(serial string) error
```

移除设备。

**参数：**
- `serial`: 设备序列号

**返回值：** 错误信息

#### Get

```go
func (m *MultiClient) Get(serial string) (*Client, bool)
```

获取指定设备的客户端。

**参数：**
- `serial`: 设备序列号

**返回值：** 客户端和是否存在

#### List

```go
func (m *MultiClient) List() []*Client
```

列出所有设备。

**返回值：** 客户端切片

#### Count

```go
func (m *MultiClient) Count() int
```

获取设备数量。

**返回值：** 设备数量

#### Broadcast

```go
func (m *MultiClient) Broadcast(msg []byte) error
```

向所有设备发送控制消息。

**参数：**
- `msg`: 控制消息

**返回值：** 错误信息

#### ForEach

```go
func (m *MultiClient) ForEach(fn func(serial string, client *Client))
```

遍历所有设备执行操作。

**参数：**
- `fn`: 回调函数

#### VideoStreamAll

```go
func (m *MultiClient) VideoStreamAll() <-chan *VideoFrameWithSerial
```

获取所有设备的视频帧通道。

**返回值：** 合并的视频帧通道

#### Close

```go
func (m *MultiClient) Close() error
```

关闭管理器。

**返回值：** 错误信息

---

### VideoFrameWithSerial 结构体

```go
type VideoFrameWithSerial struct {
    Serial string                // 设备序列号
    Frame  *video.DecodedFrame   // 视频帧
}
```

---

### WatchDevices

```go
func WatchDevices(adbClient *adb.Client, onAdd func(serial string), onRemove func(serial string)) *adb.DeviceTracker
```

监听设备上下线（便捷包装器）。

**参数：**
- `adbClient`: ADB 客户端
- `onAdd`: 设备上线回调
- `onRemove`: 设备离线回调

**返回值：** 设备跟踪器

**使用示例：**
```go
watcher := scrcpy.WatchDevices(adbClient,
    func(serial string) {
        fmt.Printf("设备上线: %s\n", serial)
    },
    func(serial string) {
        fmt.Printf("设备离线: %s\n", serial)
    },
)
ctx := context.Background()
watcher.Start(ctx)
```

#### GetDevices

```go
func GetDevices(adbClient *adb.Client) ([]types.Device, error)
```

获取已连接的设备列表。

**参数：**
- `adbClient`: ADB 客户端

**返回值：** 设备列表和错误信息

---

## pkg/adb

ADB 设备管理包，封装 ADB 命令行工具。

### Client 结构体

```go
type Client struct {
    ExecPath string // ADB 可执行文件路径，默认 "adb"
}
```

#### NewClient

```go
func NewClient(execPath string) *Client
```

创建新的 ADB 客户端。

**参数：**
- `execPath`: ADB 可执行文件路径，为空时默认 "adb"

**返回值：** Client 指针

---

### 设备管理

#### ListDevices

```go
func (c *Client) ListDevices(ctx context.Context) ([]types.Device, error)
```

列举所有已连接的 ADB 设备。

**参数：**
- `ctx`: 上下文

**返回值：** 设备列表和错误信息

#### IsDeviceConnected

```go
func (c *Client) IsDeviceConnected(ctx context.Context, serial string) (bool, error)
```

检查指定设备是否已连接。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 是否已连接和错误信息

#### GetDeviceModel

```go
func (c *Client) GetDeviceModel(ctx context.Context, serial string) (string, error)
```

获取设备型号。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 设备型号和错误信息

#### GetAndroidVersion

```go
func (c *Client) GetAndroidVersion(ctx context.Context, serial string) (int, error)
```

获取 Android 版本号。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 版本号和错误信息

#### GetDeviceProperties

```go
func (c *Client) GetDeviceProperties(ctx context.Context, serial string) (map[string]string, error)
```

获取设备所有系统属性。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 属性映射和错误信息

---

### 端口转发

#### Forward

```go
func (c *Client) Forward(ctx context.Context, serial string, remote string) (int, error)
```

建立 ADB 端口转发。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `remote`: 远程地址 (如 "localabstract:scrcpy")

**返回值：** 本地端口号和错误信息

#### RemoveForward

```go
func (c *Client) RemoveForward(ctx context.Context, serial string, port int) error
```

移除 ADB 端口转发。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `port`: 本地端口

**返回值：** 错误信息

#### RemoveAllForwards

```go
func (c *Client) RemoveAllForwards(ctx context.Context, serial string) error
```

移除指定设备的所有端口转发。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 错误信息

---

### 反向隧道

#### Reverse

```go
func (c *Client) Reverse(ctx context.Context, serial string, remoteAbstract string, localPort int) error
```

建立 ADB 反向隧道。允许 Android 设备主动连接 PC。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `remoteAbstract`: 设备上的 abstract socket 名称 (如 "scrcpy")
- `localPort`: PC 本地监听端口

**返回值：** 错误信息

#### RemoveAllReverses

```go
func (c *Client) RemoveAllReverses(ctx context.Context, serial string) error
```

移除指定设备的所有反向隧道。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 错误信息

---

### 文件操作

#### Push

```go
func (c *Client) Push(ctx context.Context, serial string, local string, remote string) error
```

推送文件到设备。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `local`: 本地文件路径
- `remote`: 设备上的目标路径

**返回值：** 错误信息

#### Pull

```go
func (c *Client) Pull(ctx context.Context, serial string, remote string, local string) error
```

从设备拉取文件。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `remote`: 设备上的文件路径
- `local`: 本地目标路径

**返回值：** 错误信息

---

### Shell

#### Shell

```go
func (c *Client) Shell(ctx context.Context, serial string, command string) (string, error)
```

在指定设备上执行 shell 命令。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `command`: shell 命令

**返回值：** 命令输出和错误信息

---

### Server 管理

#### IsServerRunning

```go
func (c *Client) IsServerRunning(ctx context.Context, serial string) (bool, error)
```

检查 scrcpy-server 是否在设备上运行。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 是否运行和错误信息

#### KillServer

```go
func (c *Client) KillServer(ctx context.Context, serial string) error
```

终止设备上的 scrcpy-server 进程。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 错误信息

---

### DeviceTracker 结构体

```go
type DeviceTracker struct {
    // 未导出字段
}
```

#### NewDeviceTracker

```go
func NewDeviceTracker(client *Client) *DeviceTracker
```

创建设备跟踪器。

**参数：**
- `client`: ADB 客户端

**返回值：** DeviceTracker 指针

#### OnAdd

```go
func (t *DeviceTracker) OnAdd(fn func(device types.Device))
```

设置设备上线回调。

**参数：**
- `fn`: 回调函数，接收 `types.Device` 参数

#### OnRemove

```go
func (t *DeviceTracker) OnRemove(fn func(device types.Device))
```

设置设备离线回调。

**参数：**
- `fn`: 回调函数，接收 `types.Device` 参数

#### OnChange

```go
func (t *DeviceTracker) OnChange(fn func(device types.Device))
```

设置设备状态变化回调。

**参数：**
- `fn`: 回调函数，接收 `types.Device` 参数

#### Start

```go
func (t *DeviceTracker) Start(ctx context.Context) error
```

启动设备跟踪。

**参数：**
- `ctx`: 上下文

**返回值：** 错误信息

#### Stop

```go
func (t *DeviceTracker) Stop()
```

停止设备跟踪。

#### IsRunning

```go
func (t *DeviceTracker) IsRunning() bool
```

检查跟踪器是否正在运行。

**返回值：** 是否运行

---

## pkg/server

Server 管理包，管理 scrcpy-server 的启动、配置和生命周期。

### 常量

```go
const (
    ServerVersion = "4.1"                    // scrcpy-server 版本号
    ServerPath    = "/data/local/tmp/scrcpy-server.jar" // 设备上的 JAR 路径
)
```

---

### Options 结构体

```go
type Options struct {
    // === 视频参数 ===
    Video        bool    // 启用视频捕获（默认 true）
    VideoCodec   string  // 编解码器: "h264", "h265", "av1", "vp8", "vp9"
    VideoSource  string  // 视频源: "display", "camera"
    VideoBitRate int     // 码率（默认 8000000）
    MaxFps       float64 // 最大帧率（0 = 不限制）
    MaxSize      int     // 最大分辨率（0 = 不限制）
    Crop         string  // 裁剪区域: "width:height:x:y"
    Angle        float64 // 旋转角度
    DisplayID    int     // 显示器 ID
    VideoEncoder string  // 指定编码器名称

    // === 音频参数 ===
    Audio        bool   // 启用音频（默认 true）
    AudioCodec   string // 编解码器: "opus", "aac", "flac", "raw"
    AudioSource  string // 音频源: "output", "mic", "playback" 等
    AudioBitRate int    // 码率（默认 128000）
    AudioEncoder string // 指定编码器
    AudioDup     bool   // 音频复制

    // === 摄像头参数 ===
    CameraID        string  // 摄像头 ID
    CameraSize      string  // 分辨率: "WxH"
    CameraFacing    string  // 方向: "front", "back", "external"
    CameraZoom      float64 // 变焦倍数
    CameraFPS       int     // 帧率
    CameraHighSpeed bool    // 高速模式
    CameraTorch     bool    // 闪光灯

    // === 控制参数 ===
    Control           bool   // 启用控制（默认 true）
    ClipboardAutosync bool   // 剪贴板自动同步（默认 true）
    ShowTouches       bool   // 显示触摸点
    StayAwake         bool   // 保持屏幕常亮
    ScreenOffTimeout  int    // 屏幕超时（ms，-1 = 不设置）
    PowerOffOnClose   bool   // 关闭时息屏
    PowerOn           bool   // 启动时亮屏（默认 true）
    KeepActive        bool   // 保持设备活跃

    // === 显示参数 ===
    NewDisplay           string // 创建虚拟显示: "WxH[/dpi]"
    CaptureOrientation   string // 捕获方向: "@0", "@1" 等
    DisplayIMEPolicy     string // 输入法策略: "local", "fallback", "hide"
    VDDestroyContent     bool   // 虚拟显示销毁内容
    VDSystemDecorations  bool   // 虚拟显示系统装饰
    FlexDisplay          bool   // 弹性显示

    // === 高级参数 ===
    TunnelForward                  bool // 使用转发隧道
    DownsizeOnError                bool // 编码出错时降分辨率（默认 true）
    Cleanup                        bool // 退出时清理（默认 true）
    IgnoreVideoEncoderConstraints  bool // 忽略编码器限制
    SendFrameMeta                  bool // 发送帧 PTS（默认 true）
    SCID                           int  // 会话客户端 ID（-1 = 不设置）

    // === 查询参数 ===
    ListEncoders    bool // 列举可用编码器
    ListDisplays    bool // 列举显示器
    ListCameras     bool // 列举摄像头
    ListCameraSizes bool // 列举摄像头分辨率
    ListApps        bool // 列举已安装应用
}
```

#### ToArgs

```go
func (o *Options) ToArgs() []string
```

将 Options 转换为 scrcpy-server 的命令行参数。

**返回值：** 参数字符串切片

---

### ServerConn 结构体

```go
type ServerConn struct {
    Serial string // 设备序列号
    Port   int    // PC 监听端口
}
```

---

### Launcher 结构体

```go
type Launcher struct {
    ADB *adb.Client // ADB 客户端
}
```

#### NewLauncher

```go
func NewLauncher(adbClient *adb.Client) *Launcher
```

创建 server 启动器。

**参数：**
- `adbClient`: ADB 客户端

**返回值：** Launcher 指针

#### Start

```go
func (l *Launcher) Start(ctx context.Context, serial string, opts Options, localJAR string, listenPort int) (*ServerConn, error)
```

启动 scrcpy-server 并建立连接。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号
- `opts`: server 启动参数
- `localJAR`: 本地 scrcpy-server.jar 路径（为空则使用设备上已有的）
- `listenPort`: PC 监听端口

**返回值：** 连接信息和错误信息

#### Kill

```go
func (l *Launcher) Kill(ctx context.Context, serial string) error
```

终止指定设备上的 scrcpy-server。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 错误信息

#### IsRunning

```go
func (l *Launcher) IsRunning(ctx context.Context, serial string) (bool, error)
```

检查 scrcpy-server 是否在运行。

**参数：**
- `ctx`: 上下文
- `serial`: 设备序列号

**返回值：** 是否运行和错误信息

---

## pkg/transport

传输层包，管理与 scrcpy-server 的三通道 socket 连接。

### Listener 结构体

Listener 管理 reverse tunnel 监听器。Android 服务端主动连接 PC。

```go
type Listener struct {
    // 未导出字段
}
```

#### NewListener

```go
func NewListener(port int) (*Listener, error)
```

创建 reverse tunnel 监听器。

**参数：**
- `port`: PC 监听端口

**返回值：** Listener 和错误信息

#### GetPort

```go
func (l *Listener) GetPort() int
```

获取监听端口（可能与输入端口不同，当 port=0 时）。

**返回值：** 实际监听端口

#### Accept

```go
func (l *Listener) Accept(video, audio, control bool, timeout time.Duration) error
```

等待 Android 服务端连接。

**参数：**
- `video`: 是否启用视频通道
- `audio`: 是否启用音频通道
- `control`: 是否启用控制通道
- `timeout`: 超时时间

**返回值：** 错误信息

#### Close

```go
func (l *Listener) Close() error
```

关闭监听器和所有连接。

**返回值：** 错误信息

---

### Connection 结构体

```go
type Connection struct {
    VideoConn   net.Conn // 视频通道连接
    AudioConn   net.Conn // 音频通道连接
    ControlConn net.Conn // 控制通道连接
}
```

#### NewConnectionFromListener

```go
func NewConnectionFromListener(l *Listener) *Connection
```

从 Listener 创建 Connection。

**参数：**
- `l`: 监听器

**返回值：** Connection 指针

#### Close

```go
func (c *Connection) Close() error
```

关闭所有通道连接。

**返回值：** 错误信息

#### IsClosed

```go
func (c *Connection) IsClosed() bool
```

检查连接是否已关闭。

**返回值：** 是否已关闭

---

### ProtocolReader 结构体

```go
type ProtocolReader struct {
    // 未导出字段
}
```

#### NewProtocolReader

```go
func NewProtocolReader(r io.Reader) *ProtocolReader
```

创建协议读取器。

**参数：**
- `r`: 读取器

**返回值：** ProtocolReader 指针

#### ReadFull

```go
func (pr *ProtocolReader) ReadFull(n int) ([]byte, error)
```

读取指定字节数的数据。

**参数：**
- `n`: 字节数

**返回值：** 数据和错误信息

#### ReadByte

```go
func (pr *ProtocolReader) ReadByte() (byte, error)
```

读取单个字节。

**返回值：** 字节和错误信息

#### ReadUint16BE

```go
func (pr *ProtocolReader) ReadUint16BE() (uint16, error)
```

读取 big-endian uint16。

**返回值：** uint16 和错误信息

#### ReadUint32BE

```go
func (pr *ProtocolReader) ReadUint32BE() (uint32, error)
```

读取 big-endian uint32。

**返回值：** uint32 和错误信息

#### ReadUint64BE

```go
func (pr *ProtocolReader) ReadUint64BE() (uint64, error)
```

读取 big-endian uint64。

**返回值：** uint64 和错误信息

#### ReadInt32BE

```go
func (pr *ProtocolReader) ReadInt32BE() (int32, error)
```

读取 big-endian int32。

**返回值：** int32 和错误信息

#### ReadInt64BE

```go
func (pr *ProtocolReader) ReadInt64BE() (int64, error)
```

读取 big-endian int64。

**返回值：** int64 和错误信息

#### ReadInt64Native

```go
func (pr *ProtocolReader) ReadInt64Native() (int64, error)
```

读取 native byte order int64。

**返回值：** int64 和错误信息

#### ReadInt32Native

```go
func (pr *ProtocolReader) ReadInt32Native() (int32, error)
```

读取 native byte order int32。

**返回值：** int32 和错误信息

#### ReadString

```go
func (pr *ProtocolReader) ReadString(n int) (string, error)
```

读取指定长度的字符串。

**参数：**
- `n`: 字符串长度

**返回值：** 字符串和错误信息

#### ReadUTF8String

```go
func (pr *ProtocolReader) ReadUTF8String() (string, error)
```

读取 UTF-8 字符串（带长度前缀）。

**返回值：** 字符串和错误信息

---

### ProtocolWriter 结构体

```go
type ProtocolWriter struct {
    // 未导出字段
}
```

#### NewProtocolWriter

```go
func NewProtocolWriter(w io.Writer) *ProtocolWriter
```

创建协议写入器。

**参数：**
- `w`: 写入器

**返回值：** ProtocolWriter 指针

#### Write

```go
func (pw *ProtocolWriter) Write(data []byte) error
```

写入数据。

**参数：**
- `data`: 数据

**返回值：** 错误信息

#### WriteByte

```go
func (pw *ProtocolWriter) WriteByte(b byte) error
```

写入单个字节。

**参数：**
- `b`: 字节

**返回值：** 错误信息

#### WriteUint16BE

```go
func (pw *ProtocolWriter) WriteUint16BE(val uint16) error
```

写入 big-endian uint16。

**参数：**
- `val`: uint16 值

**返回值：** 错误信息

#### WriteUint32BE

```go
func (pw *ProtocolWriter) WriteUint32BE(val uint32) error
```

写入 big-endian uint32。

**参数：**
- `val`: uint32 值

**返回值：** 错误信息

#### WriteUint64BE

```go
func (pw *ProtocolWriter) WriteUint64BE(val uint64) error
```

写入 big-endian uint64。

**参数：**
- `val`: uint64 值

**返回值：** 错误信息

#### WriteInt32BE

```go
func (pw *ProtocolWriter) WriteInt32BE(val int32) error
```

写入 big-endian int32。

**参数：**
- `val`: int32 值

**返回值：** 错误信息

#### WriteInt64BE

```go
func (pw *ProtocolWriter) WriteInt64BE(val int64) error
```

写入 big-endian int64。

**参数：**
- `val`: int64 值

**返回值：** 错误信息

#### WriteString

```go
func (pw *ProtocolWriter) WriteString(s string, length int) error
```

写入固定长度字符串。

**参数：**
- `s`: 字符串
- `length`: 长度

**返回值：** 错误信息

#### WriteUTF8String

```go
func (pw *ProtocolWriter) WriteUTF8String(s string) error
```

写入带 4 字节长度前缀的 UTF-8 字符串。

**参数：**
- `s`: 字符串

**返回值：** 错误信息

---

## pkg/protocol

协议编解码包，实现 scrcpy 二进制协议。

### 常量

```go
// 视频编解码器 ID
const (
    VideoCodecIDH264 uint32 = 0x68323634 // "h264"
    VideoCodecIDH265 uint32 = 0x68323635 // "h265"
    VideoCodecIDAV1  uint32 = 0x00617631 // "\0av1"
    VideoCodecIDVP8  uint32 = 0x00767038 // "\0vp8"
    VideoCodecIDVP9  uint32 = 0x00767039 // "\0vp9"
)

// 音频编解码器 ID
const (
    AudioCodecIDOpus uint32 = 0x6F707573 // "opus"
    AudioCodecIDAAC  uint32 = 0x00616163 // "\0aac"
    AudioCodecIDFLAC uint32 = 0x666C6163 // "flac"
    AudioCodecIDRAW  uint32 = 0x00726177 // "\0raw"
)

// 帧标志位
const (
    PacketFlagConfig   int64 = 1 << 62 // 配置包
    PacketFlagKeyFrame int64 = 1 << 61 // 关键帧
)

// 控制消息类型
const (
    ControlTypeInjectKeycode          byte = 0
    ControlTypeInjectText             byte = 1
    ControlTypeInjectTouchEvent       byte = 2
    ControlTypeInjectScrollEvent      byte = 3
    ControlTypeBackOrScreenOn         byte = 4
    ControlTypeExpandNotificationPanel byte = 5
    ControlTypeExpandSettingsPanel    byte = 6
    ControlTypeCollapsePanels         byte = 7
    ControlTypeGetClipboard           byte = 8
    ControlTypeSetClipboard           byte = 9
    ControlTypeSetDisplayPower        byte = 10
    ControlTypeRotateDevice           byte = 11
    ControlTypeUHIDCreate             byte = 12
    ControlTypeUHIDInput              byte = 13
    ControlTypeUHIDDestroy            byte = 14
    ControlTypeOpenHardKeyboardSettings byte = 15
    ControlTypeStartApp               byte = 16
    ControlTypeResetVideo             byte = 17
    ControlTypeCameraSetTorch         byte = 18
    ControlTypeCameraZoomIn           byte = 19
    ControlTypeCameraZoomOut          byte = 20
    ControlTypeResizeDisplay          byte = 21
    ControlTypeScanFile               byte = 22
    ControlTypeVideoSettings          byte = 101 // ws-scrcpy 扩展
    ControlTypeFilePush               byte = 102 // ws-scrcpy 扩展
)

// 设备消息类型
const (
    DeviceMsgClipboard    byte = 0
    DeviceMsgAckClipboard byte = 1
    DeviceMsgUHIDOutput   byte = 2
)

// 触摸事件动作
const (
    TouchActionDown   byte = 0
    TouchActionUp     byte = 1
    TouchActionMove   byte = 2
    TouchActionCancel byte = 3
)

// 按键事件动作
const (
    KeyActionDown  byte = 0
    KeyActionUp    byte = 1
    KeyActionMulti byte = 2
)

// 指针 ID
const (
    PointerIDMouse          uint64 = 0xFFFFFFFFFFFFFFFF
    PointerIDGenericFinger  uint64 = 0xFFFFFFFFFFFFFFFE
)

// 协议常量
const (
    DeviceNameFieldLength = 64
    MaxControlMessageSize = 1 << 18 // 256KB
    MaxClipboardTextSize  = MaxControlMessageSize - 14
    MaxTextLength         = 300
    MaxFileNameLength     = 256
    MaxFilePathLength     = 256
)
```

---

### Handshake 结构体

```go
type Handshake struct {
    DeviceName    string          // 设备型号
    Displays      []DisplayInfo   // 显示器信息列表
    Encoders      []string        // 可用编码器列表
    ClientID      int32           // 客户端 ID
}
```

#### ReadHandshake

```go
func ReadHandshake(reader *transport.ProtocolReader) (*Handshake, error)
```

从流中读取并解析握手数据。

**参数：**
- `reader`: 协议读取器

**返回值：** 握手数据和错误信息

#### GetDisplayInfo

```go
func (h *Handshake) GetDisplayInfo(displayID int32) *DisplayInfo
```

根据显示器 ID 获取显示器信息。

**参数：**
- `displayID`: 显示器 ID

**返回值：** 显示器信息

#### GetPrimaryDisplay

```go
func (h *Handshake) GetPrimaryDisplay() *DisplayInfo
```

获取主显示器 (ID=0)。

**返回值：** 显示器信息

#### GetVideoSize

```go
func (h *Handshake) GetVideoSize(displayID int32) (int32, int32)
```

获取指定显示器的视频尺寸。

**参数：**
- `displayID`: 显示器 ID

**返回值：** 宽度和高度

#### GetDeviceName

```go
func (h *Handshake) GetDeviceName() string
```

获取设备型号名称。

**返回值：** 设备名称

#### GetClientID

```go
func (h *Handshake) GetClientID() int32
```

获取客户端 ID。

**返回值：** 客户端 ID

---

### DisplayInfo 结构体

```go
type DisplayInfo struct {
    DisplayID       int32
    Width           int32
    Height          int32
    Rotation        int32
    LayerStack      int32
    Flags           int32
    ConnectionCount int32
    ScreenInfo      []byte
    VideoSettings   []byte
}
```

---

### 控制消息函数

#### InjectKeycode

```go
func InjectKeycode(action byte, keycode int32, repeat int32, metastate int32) []byte
```

构建按键注入消息 (type=0, 14 字节)。

**参数：**
- `action`: 动作 (0=DOWN, 1=UP, 2=MULTI)
- `keycode`: Android KEYCODE_* 常量
- `repeat`: 重复次数
- `metastate`: 修饰键状态

**返回值：** 消息字节切片

#### InjectText

```go
func InjectText(text string) []byte
```

构建文本注入消息 (type=1)。

**参数：**
- `text`: 要注入的文本

**返回值：** 消息字节切片

#### InjectTouchEvent

```go
func InjectTouchEvent(action byte, pointerID uint64, x, y int32, screenW, screenH uint16, pressure uint16, actionButton, buttons int32) []byte
```

构建触摸事件消息 (type=2, 32 字节)。

**参数：**
- `action`: 动作
- `pointerID`: 指针 ID
- `x, y`: 坐标
- `screenW, screenH`: 屏幕尺寸
- `pressure`: 压力值
- `actionButton`: 触摸按钮
- `buttons`: 按钮状态

**返回值：** 消息字节切片

#### InjectScrollEvent

```go
func InjectScrollEvent(x, y int32, screenW, screenH uint16, hscroll, vscroll int16, buttons int32) []byte
```

构建滚动事件消息 (type=3, 21 字节)。

**参数：**
- `x, y`: 滚动位置
- `screenW, screenH`: 屏幕尺寸
- `hscroll, vscroll`: 滚动量
- `buttons`: 按钮状态

**返回值：** 消息字节切片

#### BackOrScreenOn

```go
func BackOrScreenOn(action byte) []byte
```

构建返回键/亮屏消息 (type=4)。

**参数：**
- `action`: 动作 (0=DOWN, 1=UP)

**返回值：** 消息字节切片

#### ExpandNotificationPanel

```go
func ExpandNotificationPanel() []byte
```

构建展开通知栏消息 (type=5)。

**返回值：** 消息字节切片

#### ExpandSettingsPanel

```go
func ExpandSettingsPanel() []byte
```

构建展开设置面板消息 (type=6)。

**返回值：** 消息字节切片

#### CollapsePanels

```go
func CollapsePanels() []byte
```

构建收起面板消息 (type=7)。

**返回值：** 消息字节切片

#### GetClipboard

```go
func GetClipboard(copyKey byte) []byte
```

构建获取剪贴板消息 (type=8)。

**参数：**
- `copyKey`: 0=NONE, 1=COPY, 2=CUT

**返回值：** 消息字节切片

#### SetClipboard

```go
func SetClipboard(text string, paste bool, sequence uint64) []byte
```

构建设置剪贴板消息 (type=9)。

**参数：**
- `text`: 剪贴板内容
- `paste`: 是否自动粘贴
- `sequence`: 序列号

**返回值：** 消息字节切片

#### SetDisplayPower

```go
func SetDisplayPower(on bool) []byte
```

构建设置屏幕电源消息 (type=10)。

**参数：**
- `on`: true=亮屏, false=息屏

**返回值：** 消息字节切片

#### RotateDevice

```go
func RotateDevice() []byte
```

构建设备旋转消息 (type=11)。

**返回值：** 消息字节切片

#### UHIDCreate

```go
func UHIDCreate(id uint16, vendorID, productID uint16, name string, reportDesc []byte) []byte
```

构建创建 UHID 设备消息 (type=12)。

**参数：**
- `id`: 设备 ID
- `vendorID`: 厂商 ID
- `productID`: 产品 ID
- `name`: 设备名称
- `reportDesc`: HID 报告描述符

**返回值：** 消息字节切片

#### UHIDInput

```go
func UHIDInput(id uint16, data []byte) []byte
```

构建 UHID 输入消息 (type=13)。

**参数：**
- `id`: 设备 ID
- `data`: 输入数据

**返回值：** 消息字节切片

#### UHIDDestroy

```go
func UHIDDestroy(id uint16) []byte
```

构建销毁 UHID 设备消息 (type=14)。

**参数：**
- `id`: 设备 ID

**返回值：** 消息字节切片

#### StartApp

```go
func StartApp(packageName string) []byte
```

构建启动应用消息 (type=16)。

**参数：**
- `packageName`: 应用包名

**返回值：** 消息字节切片

#### ResetVideo

```go
func ResetVideo() []byte
```

构建重置视频消息 (type=17)。

**返回值：** 消息字节切片

#### CameraSetTorch

```go
func CameraSetTorch(on bool) []byte
```

构建设置相机闪光灯消息 (type=18)。

**参数：**
- `on`: true=开启, false=关闭

**返回值：** 消息字节切片

#### CameraZoomIn

```go
func CameraZoomIn() []byte
```

构建相机放大消息 (type=19)。

**返回值：** 消息字节切片

#### CameraZoomOut

```go
func CameraZoomOut() []byte
```

构建相机缩小消息 (type=20)。

**返回值：** 消息字节切片

#### ResizeDisplay

```go
func ResizeDisplay(width, height uint16) []byte
```

构建调整显示器大小消息 (type=21)。

**参数：**
- `width, height`: 新的显示器尺寸

**返回值：** 消息字节切片

#### ScanFile

```go
func ScanFile(path string) []byte
```

构建扫描文件消息 (type=22)。

**参数：**
- `path`: 文件路径

**返回值：** 消息字节切片

---

### 设备消息函数

#### ReadDeviceMessage

```go
func ReadDeviceMessage(reader *transport.ProtocolReader) (*DeviceMessage, error)
```

从控制通道读取设备消息。

**参数：**
- `reader`: 协议读取器

**返回值：** 设备消息和错误信息

#### ParseClipboardMessage

```go
func ParseClipboardMessage(msg *DeviceMessage) (*ClipboardMessage, error)
```

解析剪贴板消息。

**参数：**
- `msg`: 设备消息

**返回值：** 剪贴板消息和错误信息

#### ParseAckClipboardMessage

```go
func ParseAckClipboardMessage(msg *DeviceMessage) (*AckClipboardMessage, error)
```

解析剪贴板确认消息。

**参数：**
- `msg`: 设备消息

**返回值：** 剪贴板确认消息和错误信息

---

### 帧头函数

#### ReadFrameHeader

```go
func ReadFrameHeader(reader *transport.ProtocolReader) (*FrameHeader, error)
```

从流中读取帧元数据头。

**参数：**
- `reader`: 协议读取器

**返回值：** 帧头和错误信息

---

## pkg/video

视频解码包，提供视频解码和渲染功能。

### Decoder 接口

```go
type Decoder interface {
    Push(data []byte) error
    ReadFrame() (*DecodedFrame, error)
    Close() error
}
```

---

### DecodedFrame 结构体

```go
type DecodedFrame struct {
    PTS    int64  // 显示时间戳 (微秒)
    Data   []byte // 原始编码数据
    Width  int    // 帧宽度
    Height int    // 帧高度
    Config bool   // 是否为配置帧
    Key    bool   // 是否为关键帧
}
```

---

### H264Decoder 结构体

```go
type H264Decoder struct {
    // 未导出字段
}
```

#### NewH264Decoder

```go
func NewH264Decoder(capacity int) *H264Decoder
```

创建 H.264 解码器。

**参数：**
- `capacity`: 帧队列容量

**返回值：** H264Decoder 指针

#### Push

```go
func (d *H264Decoder) Push(data []byte) error
```

推送编码数据到解码器。

**参数：**
- `data`: 编码数据

**返回值：** 错误信息

#### ReadFrame

```go
func (d *H264Decoder) ReadFrame() (*DecodedFrame, error)
```

读取解码后的帧。

**返回值：** 视频帧和错误信息

#### Close

```go
func (d *H264Decoder) Close() error
```

关闭解码器。

**返回值：** 错误信息

---

### NALU 函数

#### NewNALUReader

```go
func NewNALUReader(data []byte) *NALUReader
```

创建 NALU 读取器。

**参数：**
- `data`: Annex B 格式数据

**返回值：** NALUReader 指针

#### Next

```go
func (r *NALUReader) Next() (*NALU, bool)
```

读取下一个 NALU。

**返回值：** NALU 和是否还有更多数据

#### ParseSPS

```go
func ParseSPS(data []byte) (*SPSInfo, error)
```

解析 H.264 SPS NALU。

**参数：**
- `data`: SPS 数据

**返回值：** SPS 信息和错误信息

#### HasSPSPPS

```go
func HasSPSPPS(data []byte) bool
```

检查帧数据是否包含 SPS/PPS。

**参数：**
- `data`: 帧数据

**返回值：** 是否包含

#### IsKeyFrame

```go
func IsKeyFrame(data []byte) bool
```

检查帧数据是否为关键帧。

**参数：**
- `data`: 帧数据

**返回值：** 是否为关键帧

---

## pkg/audio

音频解码包，提供音频解码和播放功能。

### Decoder 接口

```go
type Decoder interface {
    Push(data []byte) error
    ReadPCM() (*AudioData, error)
    Close() error
}
```

### Player 接口

```go
type Player interface {
    Play(pcm []byte) error
    Close() error
}
```

---

### AudioData 结构体

```go
type AudioData struct {
    PTS        int64  // 显示时间戳 (微秒)
    Data       []byte // PCM 数据
    SampleRate int    // 采样率
    Channels   int    // 声道数
    BitsPerSample int // 位深度
}
```

---

### OpusDecoder 结构体

```go
type OpusDecoder struct {
    // 未导出字段
}
```

#### NewOpusDecoder

```go
func NewOpusDecoder(capacity int) *OpusDecoder
```

创建 Opus 解码器。

**参数：**
- `capacity`: 音频包队列容量

**返回值：** OpusDecoder 指针

---

### PCMPlayer 结构体

```go
type PCMPlayer struct {
    // 未导出字段
}
```

#### NewPCMPlayer

```go
func NewPCMPlayer(capacity int) *PCMPlayer
```

创建 PCM 播放器。

**参数：**
- `capacity`: 缓冲区容量

**返回值：** PCMPlayer 指针

---

## pkg/input

输入事件包，提供触摸、键盘、滚动等输入事件构建。

### 触摸事件

#### TouchDown

```go
func TouchDown(pointerID uint64, x, y int32, screenW, screenH uint16) []byte
```

构建触摸按下事件。

**参数：**
- `pointerID`: 指针 ID
- `x, y`: 触摸坐标
- `screenW, screenH`: 屏幕尺寸

**返回值：** 消息字节切片

#### TouchMove

```go
func TouchMove(pointerID uint64, x, y int32, screenW, screenH uint16) []byte
```

构建触摸移动事件。

#### TouchUp

```go
func TouchUp(pointerID uint64, x, y int32, screenW, screenH uint16) []byte
```

构建触摸抬起事件。

#### TouchCancel

```go
func TouchCancel(pointerID uint64, x, y int32, screenW, screenH uint16) []byte
```

构建触摸取消事件。

---

### 鼠标事件

#### MouseDown

```go
func MouseDown(x, y int32, screenW, screenH uint16, button int32) []byte
```

构建鼠标按下事件。

**参数：**
- `button`: 鼠标按钮 (1=左键, 2=右键, 4=中键)

#### MouseMove

```go
func MouseMove(x, y int32, screenW, screenH uint16, buttons int32) []byte
```

构建鼠标移动事件。

#### MouseUp

```go
func MouseUp(x, y int32, screenW, screenH uint16, button int32) []byte
```

构建鼠标抬起事件。

#### Scroll

```go
func Scroll(x, y int32, screenW, screenH uint16, hscroll, vscroll int16) []byte
```

构建滚动事件。

---

### 按键事件

#### KeyDown

```go
func KeyDown(keycode int32, meta int32) []byte
```

构建按键按下事件。

**参数：**
- `keycode`: Android KEYCODE_* 常量
- `meta`: 修饰键状态

#### KeyUp

```go
func KeyUp(keycode int32, meta int32) []byte
```

构建按键抬起事件。

#### KeyPress

```go
func KeyPress(keycode int32, meta int32) [][]byte
```

构建完整按键事件 (按下 + 抬起)。

**返回值：** 两条消息的切片

#### Text

```go
func Text(text string) []byte
```

构建文本注入事件。

**参数：**
- `text`: 要注入的文本

---

### 修饰键常量

```go
const (
    MetaNone         int32 = 0
    MetaAltLeft      int32 = 0x10
    MetaAltRight     int32 = 0x20
    MetaShiftLeft    int32 = 0x40
    MetaShiftRight   int32 = 0x80
    MetaCtrlLeft     int32 = 0x1000
    MetaCtrlRight    int32 = 0x2000
    MetaMetaLeft     int32 = 0x10000
    MetaMetaRight    int32 = 0x20000
    MetaCapsLock     int32 = 0x100000
    MetaNumLock      int32 = 0x200000
    MetaScrollLock   int32 = 0x400000
)
```

---

### Android Keycode 常量

```go
const (
    KeycodeHome          int32 = 3
    KeycodeBack          int32 = 4
    KeycodeCall          int32 = 5
    KeycodeEndCall       int32 = 6
    Keycode0             int32 = 7
    // ... 完整列表见 keycode_map.go
    KeycodeA             int32 = 29
    KeycodeZ             int32 = 54
    KeycodeSpace         int32 = 62
    KeycodeEnter         int32 = 66
    KeycodeDel           int32 = 67
    KeycodeTab           int32 = 61
    KeycodeEscape        int32 = 111
    KeycodeVolumeUp      int32 = 24
    KeycodeVolumeDown    int32 = 25
    KeycodePower         int32 = 26
    KeycodeMenu          int32 = 82
    KeycodeSearch        int32 = 84
    KeycodeF1            int32 = 131
    KeycodeF12           int32 = 142
)
```

---

## pkg/control

控制功能包，提供剪贴板同步和文件推送。

### Clipboard 结构体

```go
type Clipboard struct {
    // 未导出字段
}
```

#### NewClipboard

```go
func NewClipboard(sender Sender) *Clipboard
```

创建剪贴板管理器。

**参数：**
- `sender`: 控制消息发送器

**返回值：** Clipboard 指针

#### OnChange

```go
func (c *Clipboard) OnChange(fn func(text string))
```

设置剪贴板变化回调。

**参数：**
- `fn`: 回调函数

#### HandleMessage

```go
func (c *Clipboard) HandleMessage(msgType byte, payload []byte) error
```

处理设备消息。

**参数：**
- `msgType`: 消息类型
- `payload`: 消息载荷

**返回值：** 错误信息

#### Close

```go
func (c *Clipboard) Close() error
```

关闭剪贴板管理器。

**返回值：** 错误信息

---

### FilePusher 结构体

```go
type FilePusher struct {
    // 未导出字段
}
```

#### NewFilePusher

```go
func NewFilePusher(sender Sender) *FilePusher
```

创建文件推送器。

**参数：**
- `sender`: 控制消息发送器

**返回值：** FilePusher 指针

#### HandleMessage

```go
func (f *FilePusher) HandleMessage(msgType byte, payload []byte) error
```

处理设备消息。

**参数：**
- `msgType`: 消息类型
- `payload`: 消息载荷

**返回值：** 错误信息

#### Close

```go
func (f *FilePusher) Close() error
```

关闭文件推送器。

**返回值：** 错误信息

---

### FilePushHandler 结构体

```go
type FilePushHandler struct {
    // 未导出字段
}
```

#### NewFilePushHandler

```go
func NewFilePushHandler(sender Sender) *FilePushHandler
```

创建文件推送处理器。

**参数：**
- `sender`: 控制消息发送器

**返回值：** FilePushHandler 指针

#### PushFile

```go
func (h *FilePushHandler) PushFile(filename string, data io.Reader) error
```

推送文件到设备。

**参数：**
- `filename`: 设备上的文件名
- `data`: 文件数据

**返回值：** 错误信息

#### PushFileFromPath

```go
func (h *FilePushHandler) PushFileFromPath(filename string, localPath string) error
```

从文件路径推送文件。

**参数：**
- `filename`: 设备上的文件名
- `localPath`: 本地文件路径

**返回值：** 错误信息

#### InstallAPK

```go
func (h *FilePushHandler) InstallAPK(localPath string) error
```

安装 APK 文件。

**参数：**
- `localPath`: APK 文件路径

**返回值：** 错误信息

---

## pkg/record

录制和截图包，提供屏幕录制和截图功能。

### Recorder 结构体

```go
type Recorder struct {
    // 未导出字段
}
```

#### NewRecorder

```go
func NewRecorder(output io.Writer, format string) *Recorder
```

创建录制器。

**参数：**
- `output`: 输出流
- `format`: 格式 ("mp4", "mkv")

**返回值：** Recorder 指针

#### NewRecorderFromFile

```go
func NewRecorderFromFile(path string, format string) (*Recorder, error)
```

从文件路径创建录制器。

**参数：**
- `path`: 文件路径
- `format`: 格式

**返回值：** Recorder 指针和错误信息

#### WriteVideo

```go
func (r *Recorder) WriteVideo(data []byte, pts int64, keyframe bool) error
```

写入视频帧。

**参数：**
- `data`: 编码数据
- `pts`: 显示时间戳
- `keyframe`: 是否为关键帧

**返回值：** 错误信息

#### WriteAudio

```go
func (r *Recorder) WriteAudio(data []byte, pts int64) error
```

写入音频数据。

**参数：**
- `data`: 编码数据
- `pts`: 显示时间戳

**返回值：** 错误信息

#### GetFrameCount

```go
func (r *Recorder) GetFrameCount() int64
```

获取已录制帧数。

**返回值：** 帧数

#### Close

```go
func (r *Recorder) Close() error
```

关闭录制器。

**返回值：** 错误信息

---

### Screenshot 结构体

```go
type Screenshot struct {
    // 未导出字段
}
```

#### NewScreenshot

```go
func NewScreenshot(width, height int) *Screenshot
```

创建截图器。

**参数：**
- `width`: 宽度
- `height`: 高度

**返回值：** Screenshot 指针

#### Capture

```go
func (s *Screenshot) Capture(data []byte, format string) ([]byte, error)
```

截取当前帧。

**参数：**
- `data`: 视频帧数据
- `format`: 输出格式 ("png", "jpeg")

**返回值：** 图片数据和错误信息

#### SaveToFile

```go
func (s *Screenshot) SaveToFile(data []byte, path string, format string) error
```

保存截图到文件。

**参数：**
- `data`: 图片数据
- `path`: 文件路径
- `format`: 格式

**返回值：** 错误信息

#### SetSize

```go
func (s *Screenshot) SetSize(width, height int)
```

设置截图尺寸。

**参数：**
- `width`: 宽度
- `height`: 高度

#### GetSize

```go
func (s *Screenshot) GetSize() (int, int)
```

获取截图尺寸。

**返回值：** 宽度和高度

---

## pkg/types

公共类型包，定义跨包使用的类型。

### Device 结构体

```go
type Device struct {
    Serial    string // 设备序列号
    State     string // 设备状态
    Model     string // 设备型号
    Product   string // 产品名称
    Transport string // 传输方式
}
```

### DeviceEvent 结构体

```go
type DeviceEvent struct {
    Type   EventType // 事件类型
    Device Device    // 关联的设备信息
}
```

### EventType 常量

```go
const (
    EventDeviceAdded   EventType = iota // 设备上线
    EventDeviceRemoved                  // 设备离线
    EventDeviceChanged                  // 设备状态变化
)
```

### DeviceInfo 结构体

```go
type DeviceInfo struct {
    Serial         string            // 设备序列号
    Model          string            // 设备型号
    AndroidVersion int               // Android 版本号
    SDKVersion     int               // SDK 版本号
    Properties     map[string]string // 设备属性
}
```

### DisplayInfo 结构体

```go
type DisplayInfo struct {
    DisplayID  int32 // 显示器 ID
    Width      int32 // 宽度
    Height     int32 // 高度
    Rotation   int32 // 旋转
    LayerStack int32 // 图层栈
    Flags      int32 // 标志位
}
```

### ScreenInfo 结构体

```go
type ScreenInfo struct {
    ContentRect    Rect  // 内容区域
    VideoSize      Size  // 视频尺寸
    DeviceRotation uint8 // 设备旋转
}
```

### Rect 结构体

```go
type Rect struct {
    Left   int32
    Top    int32
    Right  int32
    Bottom int32
}
```

### Size 结构体

```go
type Size struct {
    Width  int32
    Height int32
}
```

### VideoFrame 结构体

```go
type VideoFrame struct {
    PTS    int64  // 显示时间戳
    Data   []byte // 原始编码数据
    Width  int    // 帧宽度
    Height int    // 帧高度
    Config bool   // 是否为配置帧
    Key    bool   // 是否为关键帧
}
```

### AudioPacket 结构体

```go
type AudioPacket struct {
    PTS  int64  // 显示时间戳
    Data []byte // 编码数据
}
```

### VideoCodec 枚举

```go
const (
    VideoCodecH264 VideoCodec = iota
    VideoCodecH265
    VideoCodecAV1
    VideoCodecVP8
    VideoCodecVP9
)
```

### AudioCodec 枚举

```go
const (
    AudioCodecOpus AudioCodec = iota
    AudioCodecAAC
    AudioCodecFLAC
    AudioCodecRAW
)
```