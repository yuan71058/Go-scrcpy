# Go-scrcpy 二进制协议规范

本文档描述 scrcpy-server v4.1 的完整通信协议。

## 连接架构

```
Go Client                          Android Server
    │                                     │
    │←─── Video Socket (encoded H264) ────│
    │←─── Audio Socket (encoded Opus) ────│
    │──── Control Socket (bidirectional) ─→│
    │                                     │
```

三通道通过 ADB 端口转发建立连接。

---

## 连接流程

### Step 1: 启动 Server

```
adb push scrcpy-server.jar /data/local/tmp/
adb shell CLASSPATH=/data/local/tmp/scrcpy-server.jar \
  nohup app_process / com.genymobile.scrcpy.Server \
  <version> [key=value ...]
```

### Step 2: ADB 端口转发

```bash
adb forward tcp:<localPort1> localabstract:scrcpy    # video
adb forward tcp:<localPort2> localabstract:scrcpy_1  # audio (auto-assigned)
adb forward tcp:<localPort3> localabstract:scrcpy_2  # control (auto-assigned)
```

### Step 3: 握手

#### 3a. Dummy Byte (可选)

如果 `send_dummy_byte=true`，server 在第一个连接的 socket 上写入 1 字节 `0x00`。客户端可据此检测连接错误。

#### 3b. 设备元数据

Server 发送 **64 字节**设备名：

```
[64 bytes] 设备型号 (UTF-8, null 填充到 64 字节)
```

#### 3c. 流头（每个 stream）

**视频流头 - 4 字节 Codec ID:**

| Codec | Hex | ASCII |
|-------|-----|-------|
| H.264 | `0x68323634` | `h264` |
| H.265 | `0x68323635` | `h265` |
| AV1   | `0x00617631` | `\0av1` |
| VP8   | `0x00767038` | `\0vp8` |
| VP9   | `0x00767039` | `\0vp9` |

**音频流头 - 4 字节 Codec ID:**

| Codec | Hex | ASCII |
|-------|-----|-------|
| Opus  | `0x6F707573` | `opus` |
| AAC   | `0x00616163` | `\0aac` |
| FLAC  | `0x666C6163` | `flac` |
| RAW   | `0x00726177` | `\0raw` |

#### 3d. 会话元数据

每次捕获会话开始或 resize 后发送：

```
[4B] Flags (int32 BE):
     bit 63 = PACKET_FLAG_SESSION
     bit 0  = isClientResize
[4B] Width (int32 BE)
[4B] Height (int32 BE)
```

#### 3e. 流禁用通知

如果音频无法捕获或配置错误：

```
[4B] 全零    → code 0: 音频禁用（继续）
[4B] byte[3]=1 → code 1: 致命配置错误（客户端必须停止）
```

---

## 帧数据格式

### 帧元数据（per-packet header, 12B）

当 `send_frame_meta=true` 时，每个编码包前附带：

```
[8B] PTS + flags (int64, native byte order):
     bit 62 = PACKET_FLAG_CONFIG (SPS/PPS 配置包)
     bit 61 = 关键帧标记
     其余位 = PTS 微秒时间戳
[4B] Packet size (int32, native byte order)
```

### 视频帧

```
[12B] Frame header (if send_frame_meta)
[NB]  Encoded NAL units
```

NALU 使用 Annex B 格式（start code `00 00 00 01`）。

### 音频帧

```
[12B] Frame header (if send_frame_meta)
[NB]  Encoded audio data (Opus/AAC/FLAC/PCM frame)
```

---

## 控制消息格式 (Client → Server)

所有整数 **big-endian**。首字节为消息类型。

### 编码辅助

```
Position (12 bytes):
  [0..3]   int32be   x
  [4..7]   int32be   y
  [8..9]   uint16be  screenWidth
  [10..11] uint16be  screenHeight

String_4 (4 字节长度前缀):
  [0..3]   uint32be  length
  [4..len] UTF-8 data

String_1 (1 字节长度前缀):
  [0]      uint8     length
  [1..len] UTF-8 data

u16fp: unsigned 16-bit fixed-point [0,1] → f * 0x10000
i16fp: signed 16-bit fixed-point [-1,1] → f * 0x8000
```

### 消息类型表

#### 0 — INJECT_KEYCODE (14 bytes)

```
[0]      uint8     type = 0
[1]      uint8     action      (0=DOWN, 1=UP, 2=MULTI)
[2..5]   int32be   keycode     (Android KEYCODE_*)
[6..9]   int32be   repeat
[10..13] int32be   metastate   (META_* bitmask)
```

#### 1 — INJECT_TEXT (5 + N bytes)

```
[0]      uint8     type = 1
[1..4]   uint32be  text_length (max 300)
[5..len] UTF-8     text
```

#### 2 — INJECT_TOUCH_EVENT (32 bytes)

```
[0]       uint8     type = 2
[1]       uint8     action       (AMOTION_EVENT_ACTION_*)
[2..9]    uint64be  pointer_id   (UINT64_MAX=mouse, UINT64_MAX-1=generic finger)
[10..11]  int32be   x            (Position)
[12..15]  int32be   y            (Position)
[16..17]  uint16be  screenWidth  (Position)
[18..19]  uint16be  screenHeight (Position)
[20..21]  uint16be  pressure     (u16fp: 0x0000=0.0, 0xFFFF=1.0)
[22..25]  int32be   action_button (BUTTON_*)
[26..29]  int32be   buttons      (BUTTON_*)
```

**Action 常量:**
- `0` = DOWN
- `1` = UP
- `2` = MOVE
- `3` = CANCEL

**Pointer ID:**
- `0xFFFFFFFFFFFFFFFF` = 鼠标
- `0xFFFFFFFFFFFFFFFE` = 通用手指

#### 3 — INJECT_SCROLL_EVENT (21 bytes)

```
[0]       uint8     type = 3
[1..4]    int32be   x
[5..8]    int32be   y
[9..10]   uint16be  screenWidth
[11..12]  uint16be  screenHeight
[13..14]  uint16be  hscroll     (i16fp * 16, 范围 [-16, 16])
[15..16]  uint16be  vscroll     (i16fp * 16, 范围 [-16, 16])
[17..20]  int32be   buttons
```

#### 4 — BACK_OR_SCREEN_ON (2 bytes)

```
[0]  uint8  type = 4
[1]  uint8  action  (0=DOWN, 1=UP)
```

#### 5 — EXPAND_NOTIFICATION_PANEL (1 byte)

```
[0]  uint8  type = 5
```

#### 6 — EXPAND_SETTINGS_PANEL (1 byte)

```
[0]  uint8  type = 6
```

#### 7 — COLLAPSE_PANELS (1 byte)

```
[0]  uint8  type = 7
```

#### 8 — GET_CLIPBOARD (2 bytes)

```
[0]  uint8  type = 8
[1]  uint8  copy_key  (0=NONE, 1=COPY, 2=CUT)
```

#### 9 — SET_CLIPBOARD (10 + N bytes)

```
[0]       uint8     type = 9
[1..8]    uint64be  sequence    (单调递增, 0=无效/无同步)
[9]       uint8     paste       (0 or 1)
[10..13]  uint32be  text_length (max 262133 = 256K - 14)
[14..len] UTF-8     text
```

#### 10 — SET_DISPLAY_POWER (2 bytes)

```
[0]  uint8  type = 10
[1]  uint8  on  (0=off, 1=on)
```

#### 11 — ROTATE_DEVICE (1 byte)

```
[0]  uint8  type = 11
```

#### 12 — UHID_CREATE (9 + N bytes)

```
[0]       uint8     type = 12
[1..2]    uint16be  id
[3..4]    uint16be  vendor_id
[5..6]    uint16be  product_id
[7]       uint8     name_length (max 127)
[8..len]  UTF-8     name
[len+1..len+2] uint16be report_desc_size
[len+3..end]   uint8[]  report_desc (raw HID descriptor)
```

#### 13 — UHID_INPUT (5 + N bytes)

```
[0]       uint8     type = 13
[1..2]    uint16be  id
[3..4]    uint16be  size (max SC_HID_MAX_SIZE)
[5..len]  uint8[]   data
```

#### 14 — UHID_DESTROY (3 bytes)

```
[0]  uint8  type = 14
[1..2] uint16be  id
```

#### 15 — OPEN_HARD_KEYBOARD_SETTINGS (1 byte)

```
[0]  uint8  type = 15
```

#### 16 — START_APP (2 + N bytes)

```
[0]       uint8     type = 16
[1]       uint8     name_length (max 255)
[2..len]  UTF-8     package_name
```

#### 17 — RESET_VIDEO (1 byte)

```
[0]  uint8  type = 17
```

#### 18 — CAMERA_SET_TORCH (2 bytes)

```
[0]  uint8  type = 18
[1]  uint8  on  (0=off, 1=on)
```

#### 19 — CAMERA_ZOOM_IN (1 byte)

```
[0]  uint8  type = 19
```

#### 20 — CAMERA_ZOOM_OUT (1 byte)

```
[0]  uint8  type = 20
```

#### 21 — RESIZE_DISPLAY (5 bytes)

```
[0]       uint8     type = 21
[1..2]    uint16be  width
[3..4]    uint16be  height
```

#### 22 — SCAN_FILE (5 + N bytes)

```
[0]       uint8     type = 22
[1..4]    uint32be  path_length (max 256)
[5..len]  UTF-8     path
```

#### 101 — VIDEO_SETTINGS (ws-scrcpy 扩展, 35+ bytes)

```
[0]       uint8     type = 101
[1..4]    int32be   bitrate
[5..8]    int32be   maxFps
[9]       int8      iFrameInterval
[10..11]  int16be   bounds.width  (0 = no bound)
[12..13]  int16be   bounds.height
[14..15]  int16be   crop.left
[16..17]  int16be   crop.top
[18..19]  int16be   crop.right
[20..21]  int16be   crop.bottom
[22]      int8      sendFrameMeta (0 or 1)
[23]      int8      lockedVideoOrientation (-1 = unlocked)
[24..27]  int32be   displayId
[28..31]  int32be   codecOptions string length
[32..31+N]  bytes   codecOptions UTF-8 (if length > 0)
[...]     int32be   encoderName string length
[...]     bytes     encoderName UTF-8 (if length > 0)
```

#### 102 — FILE_PUSH (ws-scrcpy 扩展, 变长)

子命令:
```
[0]       uint8     type = 102
[1]       uint8     sub_type:
                      0 = NEW
                      1 = START
                      2 = APPEND
                      3 = FINISH
```

---

## 设备消息格式 (Server → Client)

### 控制通道消息

所有设备消息通过控制通道返回。

#### 0 — CLIPBOARD

```
[0]      uint8     type = 0
[1..4]   uint32be  text_length
[5..len] UTF-8     text
```

#### 1 — ACK_CLIPBOARD

```
[0]      uint8     type = 1
[1..8]   uint64be  sequence
```

#### 2 — UHID_OUTPUT

```
[0]      uint8     type = 2
[1..2]   uint16be  id
[3..4]   uint16be  data_length
[5..len] uint8[]   data
```

### ws-scrcpy 扩展消息

#### 101 — PUSH_RESPONSE

```
[0]      uint8     type = 101
[1..2]   int16be   push_id
[3]      uint8     code
```

---

## DisplayInfo 结构 (24 bytes)

每个 display 的元数据，在握手阶段发送：

```
[0..3]   int32be   displayId
[4..7]   int32be   width
[8..11]  int32be   height
[12..15] int32be   rotation    (0, 1, 2, 3)
[16..19] int32be   layerStack
[20..23] int32be   flags
```

**Flags 常量:**
- `0x001` = FLAG_SUPPORTS_PROTECTED_BUFFERS
- `0x002` = FLAG_SECURE
- `0x004` = FLAG_PRIVATE
- `0x008` = FLAG_PRESENTATION
- `0x010` = FLAG_ROUND

## ScreenInfo 结构 (25 bytes)

```
[0..3]   int32be   contentRect.left
[4..7]   int32be   contentRect.top
[8..11]  int32be   contentRect.right
[12..15] int32be   contentRect.bottom
[16..19] int32be   videoSize.width
[20..23] int32be   videoSize.height
[24]     uint8     deviceRotation
```

---

## 音频源列表

| 名称 | CLI 值 | Android API |
|------|--------|-------------|
| OUTPUT | `output` | REMOTE_SUBMIX |
| PLAYBACK | `playback` | AudioPlaybackCapture (需 Android 13+) |
| MIC | `mic` | MIC |
| MIC_UNPROCESSED | `mic-unprocessed` | UNPROCESSED |
| MIC_CAMCORDER | `mic-camcorder` | CAMCORDER |
| MIC_VOICE_RECOGNITION | `mic-voice-recognition` | VOICE_RECOGNITION |
| MIC_VOICE_COMMUNICATION | `mic-voice-communication` | VOICE_COMMUNICATION |
| VOICE_CALL | `voice-call` | VOICE_CALL |
| VOICE_CALL_UPLINK | `voice-call-uplink` | VOICE_UPLINK |
| VOICE_CALL_DOWNLINK | `voice-call-downlink` | VOICE_DOWNLINK |
| VOICE_PERFORMANCE | `voice-performance` | VOICE_PERFORMANCE |

---

## 关键常量

```go
const (
    DeviceNameFieldLength = 64
    MaxControlMessageSize = 1 << 18  // 256KB
    MaxClipboardTextSize  = MaxControlMessageSize - 14
    MaxTextLength         = 300
    MaxFileNameLength     = 256
    MaxFilePathLength     = 256
)
```

---

## 多客户端协调

当多个客户端连接同一设备的 scrcpy-server 时：

1. Server 在握手消息中包含 `connectionCount`
2. 新客户端应检查已有客户端的 VideoSettings bounds
3. 新客户端的 bounds 不应超过已有客户端
4. 通过 `VideoSettings` (type=101) 动态调整编码参数
5. 所有客户端共享同一个编码器输出流

---

## 注意事项

1. 所有多字节整数为 **big-endian**
2. 帧元数据中的 PTS 使用 **native byte order**（通常 little-endian on Android/ARM）
3. 消息最大 256KB
4. `UHID_CREATE` 和 `UHID_DESTROY` 不可从队列丢弃
5. `SET_CLIPBOARD` 的 sequence 必须单调递增
6. 触摸事件的 pressure 使用 u16fp 定点数格式
7. 滚动事件的 scroll 值乘以 16 后编码为 i16fp
