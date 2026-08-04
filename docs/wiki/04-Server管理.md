# Go-scrcpy Code Wiki — Server 管理（pkg/server）

`pkg/server` 管理 `scrcpy-server.jar` 在 Android 设备上的生命周期，包括启动、停止和参数序列化。

## 包结构

```
pkg/server/
├── version.go    # 版本常量
├── options.go    # Options 结构体 + ToArgs() 序列化
└── launcher.go   # Launcher — 启动/停止 server
```

## 常量

```go
const (
    ServerVersion = "4.1"                              // scrcpy-server 版本号
    ServerPath    = "/data/local/tmp/scrcpy-server.jar" // 设备上的 JAR 路径
    PIDFile       = "/data/local/tmp/go_scrcpy.pid"    // PID 文件路径
)
```

## 类型定义

### Options — Server 启动参数

```go
type Options struct {
    // === 视频参数 ===
    Video        bool      // 启用视频捕获（默认 true）
    VideoCodec   string    // 编解码器: "h264", "h265", "av1", "vp8", "vp9"
    VideoSource  string    // 视频源: "display", "camera"
    VideoBitRate int       // 码率（默认 8000000）
    MaxFps       float64   // 最大帧率（0 = 不限制）
    MaxSize      int       // 最大分辨率（0 = 不限制）
    Crop         string    // 裁剪区域: "width:height:x:y"
    Angle        float64   // 旋转角度
    DisplayID    int       // 显示器 ID
    VideoEncoder string    // 指定编码器名称

    // === 音频参数 ===
    Audio        bool      // 启用音频（默认 true）
    AudioCodec   string    // 编解码器: "opus", "aac", "flac", "raw"
    AudioSource  string    // 音频源: "output", "mic", "playback" 等
    AudioBitRate int       // 码率（默认 128000）
    AudioEncoder string    // 指定编码器
    AudioDup     bool      // 音频复制

    // === 摄像头参数 ===
    CameraID        string  // 摄像头 ID
    CameraSize      string  // 分辨率: "WxH"
    CameraFacing    string  // 方向: "front", "back", "external"
    CameraZoom      float64 // 变焦倍数
    CameraFPS       int     // 帧率
    CameraHighSpeed bool    // 高速模式
    CameraTorch     bool    // 闪光灯

    // === 控制参数 ===
    Control           bool  // 启用控制（默认 true）
    ClipboardAutosync bool  // 剪贴板自动同步（默认 true）
    ShowTouches       bool  // 显示触摸点
    StayAwake         bool  // 保持屏幕常亮
    ScreenOffTimeout  int   // 屏幕超时（ms，-1 = 不设置）
    PowerOffOnClose   bool  // 关闭时息屏
    PowerOn           bool  // 启动时亮屏（默认 true）
    KeepActive        bool  // 保持设备活跃

    // === 显示参数 ===
    NewDisplay           string  // 创建虚拟显示: "WxH[/dpi]"
    CaptureOrientation   string  // 捕获方向
    DisplayIMEPolicy     string  // 输入法策略
    VDDestroyContent     bool    // 虚拟显示销毁内容
    VDSystemDecorations  bool    // 虚拟显示系统装饰
    FlexDisplay          bool    // 弹性显示

    // === 高级参数 ===
    TunnelForward                  bool  // 使用转发隧道
    DownsizeOnError                bool  // 编码出错时降分辨率（默认 true）
    Cleanup                        bool  // 退出时清理（默认 true）
    IgnoreVideoEncoderConstraints  bool  // 忽略编码器限制
    SendFrameMeta                  bool  // 发送帧 PTS（默认 true）
    SCID                           int   // 会话客户端 ID（-1 = 不设置）

    // === 查询参数 ===
    ListEncoders    bool  // 列举可用编码器
    ListDisplays    bool  // 列举显示器
    ListCameras     bool  // 列举摄像头
    ListCameraSizes bool  // 列举摄像头分辨率
    ListApps        bool  // 列举已安装应用
}
```

**ToArgs() 方法：** 将 Options 序列化为 scrcpy-server 的命令行参数数组。

```go
func (o *Options) ToArgs() []string
```

输出格式示例：`["4.1", "video=true", "audio=true", "video_bit_rate=8000000", ...]`

序列化规则：
- 仅当字段值**非零值**时才添加对应参数
- `Video=true` 是默认值，仅在 `Video=false` 时添加
- 浮点数使用 `formatFloat()` 去除末尾多余的零

### ServerConn — 连接信息

```go
type ServerConn struct {
    Serial string  // 设备序列号
    Port   int     // PC 监听端口
}
```

### Launcher — Server 启动器

```go
type Launcher struct {
    ADB *adb.Client  // ADB 客户端
}
```

**构造函数：**

```go
func NewLauncher(adbClient *adb.Client) *Launcher
```

**核心方法：**

| 方法 | 说明 |
|------|------|
| `Start(ctx, serial, opts, localJAR, listenPort)` | 启动 server 并建立连接 |
| `Kill(ctx, serial)` | 终止 server |
| `IsRunning(ctx, serial)` | 检查 server 是否运行 |

**Start() 启动流程：**

```
1. 检查 server 是否已在运行
   └── 如果运行中，先 Kill 旧进程
2. 推送 JAR 文件（如果 localJAR 非空）
   └── adb push <localJAR> /data/local/tmp/scrcpy-server.jar
3. 设置 reverse tunnel
   └── adb reverse localabstract:scrcpy tcp:<listenPort>
4. 构建启动命令并执行
   └── CLASSPATH=/data/local/tmp/scrcpy-server.jar \
       nohup app_process / com.genymobile.scrcpy.Server \
       <version> [key=value ...] \
       >/data/local/tmp/scrcpy-server.log 2>&1 &
5. 等待 server 启动（固定 2 秒延迟）
6. 返回 ServerConn{Serial, Port}
```

**Kill() 清理流程：**

```
1. adb reverse --remove-all     // 移除所有反向隧道
2. adb shell pkill -f scrcpy   // 终止 server 进程
```

## 日志系统

独立的日志级别控制，日志前缀 `[SERVER DEBUG]`, `[SERVER INFO]`, `[SERVER ERROR]`。