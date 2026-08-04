# Go-scrcpy Code Wiki — ADB 管理（pkg/adb）

`pkg/adb` 封装 ADB 命令行工具，提供设备发现、设备跟踪、端口转发、文件操作和 server 进程管理功能。

## 包结构

```
pkg/adb/
├── client.go    # ADB Client 核心 — 命令执行、设备列举、属性查询
├── device.go    # DeviceTracker — 实时设备上下线跟踪
└── forward.go   # 端口转发、反向隧道、文件推送/拉取、server 管理
```

## 类型定义

### Client — ADB 客户端

```go
type Client struct {
    ExecPath string  // ADB 可执行文件路径，默认 "adb"
    mu       sync.Mutex
}
```

**构造函数：**

```go
func NewClient(execPath string) *Client
```

`execPath` 为空时默认使用系统 PATH 中的 `"adb"`。

### 核心方法

#### 命令执行（内部方法）

```go
func (c *Client) runCommand(ctx, args...) (string, error)
func (c *Client) runDeviceCommand(ctx, serial, args...) (string, error)
```

- `runCommand` — 执行 ADB 全局命令（如 `adb devices`）
- `runDeviceCommand` — 执行针对指定设备的命令（自动添加 `-s serial`）
- 使用 `sync.Mutex` 保证并发安全
- 使用 `exec.CommandContext` 支持上下文取消

#### 设备管理

| 方法 | 说明 | 底层命令 |
|------|------|----------|
| `ListDevices(ctx)` | 列举所有已连接设备 | `adb devices -l` |
| `IsDeviceConnected(ctx, serial)` | 检查设备是否连接 | `adb -s <serial> get-state` |
| `GetDeviceModel(ctx, serial)` | 获取设备型号 | `adb shell getprop ro.product.model` |
| `GetAndroidVersion(ctx, serial)` | 获取 Android SDK 版本 | `adb shell getprop ro.build.version.sdk` |
| `GetDeviceProperties(ctx, serial)` | 获取所有系统属性 | `adb shell getprop` |
| `Shell(ctx, serial, command)` | 执行 shell 命令 | `adb shell <command>` |

**`ListDevices` 输出解析：**

`adb devices -l` 输出格式：
```
List of devices attached
emulator-5554          device product:sdk_gphone64_arm64 model:emu64a transport_id:1
```

`parseDeviceLine()` 函数解析每行，提取 `Serial`、`State`、`Product`、`Model`、`Transport`。

**`cleanSerial()` 函数：** 清理 Windows 上 `adb track-devices` 可能给序列号添加的垃圾前缀（如 `"0055127.0.0.1:5555"` → `"127.0.0.1:5555"`）。

#### 端口转发

| 方法 | 说明 | 底层命令 |
|------|------|----------|
| `Forward(ctx, serial, remote)` | 建立端口转发，返回本地端口 | `adb forward tcp:<port> <remote>` |
| `RemoveForward(ctx, serial, port)` | 移除端口转发 | `adb forward --remove tcp:<port>` |
| `RemoveAllForwards(ctx, serial)` | 移除所有转发 | `adb forward --remove-all` |
| `Reverse(ctx, serial, abstract, port)` | 建立反向隧道 | `adb reverse localabstract:<name> tcp:<port>` |
| `RemoveAllReverses(ctx, serial)` | 移除所有反向隧道 | `adb reverse --remove-all` |

**`Forward` 的智能复用：** 先检查是否已有到同一 `remote` 地址的转发，有则复用，无则分配新端口。

**`findAvailablePort()`：** 通过监听 `127.0.0.1:0` 获取系统分配的可用端口。

#### 文件操作

| 方法 | 说明 | 底层命令 |
|------|------|----------|
| `Push(ctx, serial, local, remote)` | 推送文件到设备 | `adb push <local> <remote>` |
| `Pull(ctx, serial, remote, local)` | 从设备拉取文件 | `adb pull <remote> <local>` |

#### Server 进程管理

| 方法 | 说明 | 底层命令 |
|------|------|----------|
| `IsServerRunning(ctx, serial)` | 检查 scrcpy-server 是否运行 | `adb shell ps -A \| grep scrcpy` |
| `KillServer(ctx, serial)` | 终止 scrcpy-server | `adb shell pkill -f scrcpy` |

### DeviceTracker — 设备跟踪器

```go
type DeviceTracker struct {
    client     *Client
    onAdd      func(device types.Device)
    onRemove   func(device types.Device)
    onChange   func(device types.Device)
    mu         sync.RWMutex
    running    bool
    cancelFunc context.CancelFunc
}
```

**构造函数：**

```go
func NewDeviceTracker(client *Client) *DeviceTracker
```

**回调设置方法：**

| 方法 | 说明 |
|------|------|
| `OnAdd(fn)` | 设备上线回调 |
| `OnRemove(fn)` | 设备离线回调 |
| `OnChange(fn)` | 设备状态变化回调 |

**生命周期方法：**

| 方法 | 说明 |
|------|------|
| `Start(ctx)` | 启动跟踪（通过 `adb track-devices -l` 实时监听） |
| `Stop()` | 停止跟踪 |
| `IsRunning()` | 检查是否运行中 |

**工作原理：**

1. 执行 `adb track-devices -l` 命令
2. 获取 stdout 管道，在 goroutine 中逐行读取
3. 维护 `knownDevices` 映射，与每次新输出对比差异
4. 发现新设备 → 调用 `onAdd` 回调
5. 设备消失 → 调用 `onRemove` 回调
6. 状态变化 → 调用 `onChange` 回调
7. 启动时先获取当前设备列表，触发 `onAdd` 处理已有设备

## 日志系统

独立的日志级别控制（与 `pkg/scrcpy` 相同的日志级别常量体系）：

```go
func SetLogLevel(level int)
```

日志前缀：`[ADB DEBUG]`, `[ADB INFO]`, `[ADB ERROR]`