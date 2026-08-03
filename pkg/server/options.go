package server

import (
	"fmt"
	"strconv"
	"strings"
)

// Options 定义 scrcpy-server 的启动参数
// 零值即为合理默认值
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
	// VideoCodecOptions map[string]string // 编码器参数（预留，暂不实现）

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

// ToArgs 将 Options 转换为 scrcpy-server 的命令行参数
// 返回格式: ["4.1", "video=true", "audio=true", ...]
func (o *Options) ToArgs() []string {
	args := []string{ServerVersion}

	// 视频参数
	if !o.Video {
		args = append(args, "video=false")
	}
	if o.VideoCodec != "" {
		args = append(args, "video_codec="+o.VideoCodec)
	}
	if o.VideoSource != "" {
		args = append(args, "video_source="+o.VideoSource)
	}
	if o.VideoBitRate > 0 {
		args = append(args, "video_bit_rate="+strconv.Itoa(o.VideoBitRate))
	}
	if o.MaxFps > 0 {
		args = append(args, "max_fps="+formatFloat(o.MaxFps))
	}
	if o.MaxSize > 0 {
		args = append(args, "max_size="+strconv.Itoa(o.MaxSize))
	}
	if o.Crop != "" {
		args = append(args, "crop="+o.Crop)
	}
	if o.Angle != 0 {
		args = append(args, "angle="+formatFloat(o.Angle))
	}
	if o.DisplayID > 0 {
		args = append(args, "display_id="+strconv.Itoa(o.DisplayID))
	}
	if o.VideoEncoder != "" {
		args = append(args, "video_encoder="+o.VideoEncoder)
	}

	// 音频参数
	if !o.Audio {
		args = append(args, "audio=false")
	}
	if o.AudioCodec != "" {
		args = append(args, "audio_codec="+o.AudioCodec)
	}
	if o.AudioSource != "" {
		args = append(args, "audio_source="+o.AudioSource)
	}
	if o.AudioBitRate > 0 {
		args = append(args, "audio_bit_rate="+strconv.Itoa(o.AudioBitRate))
	}
	if o.AudioEncoder != "" {
		args = append(args, "audio_encoder="+o.AudioEncoder)
	}
	if o.AudioDup {
		args = append(args, "audio_dup=true")
	}

	// 摄像头参数
	if o.CameraID != "" {
		args = append(args, "camera_id="+o.CameraID)
	}
	if o.CameraSize != "" {
		args = append(args, "camera_size="+o.CameraSize)
	}
	if o.CameraFacing != "" {
		args = append(args, "camera_facing="+o.CameraFacing)
	}
	if o.CameraZoom > 0 {
		args = append(args, "camera_zoom="+formatFloat(o.CameraZoom))
	}
	if o.CameraFPS > 0 {
		args = append(args, "camera_fps="+strconv.Itoa(o.CameraFPS))
	}
	if o.CameraHighSpeed {
		args = append(args, "camera_high_speed=true")
	}
	if o.CameraTorch {
		args = append(args, "camera_torch=true")
	}

	// 控制参数
	if !o.Control {
		args = append(args, "control=false")
	}
	if !o.ClipboardAutosync {
		args = append(args, "clipboard_autosync=false")
	}
	if o.ShowTouches {
		args = append(args, "show_touches=true")
	}
	if o.StayAwake {
		args = append(args, "stay_awake=true")
	}
	if o.ScreenOffTimeout >= 0 {
		args = append(args, "screen_off_timeout="+strconv.Itoa(o.ScreenOffTimeout))
	}
	if o.PowerOffOnClose {
		args = append(args, "power_off_on_close=true")
	}
	if !o.PowerOn {
		args = append(args, "power_on=false")
	}
	if o.KeepActive {
		args = append(args, "keep_active=true")
	}

	// 显示参数
	if o.NewDisplay != "" {
		args = append(args, "new_display="+o.NewDisplay)
	}
	if o.CaptureOrientation != "" {
		args = append(args, "capture_orientation="+o.CaptureOrientation)
	}
	if o.DisplayIMEPolicy != "" {
		args = append(args, "display_ime_policy="+o.DisplayIMEPolicy)
	}
	if o.VDDestroyContent {
		args = append(args, "vd_destroy_content=true")
	}
	if o.VDSystemDecorations {
		args = append(args, "vd_system_decorations=true")
	}
	if o.FlexDisplay {
		args = append(args, "flex_display=true")
	}

	// 高级参数
	if o.TunnelForward {
		args = append(args, "tunnel_forward=true")
	}
	if !o.DownsizeOnError {
		args = append(args, "downsize_on_error=false")
	}
	if !o.Cleanup {
		args = append(args, "cleanup=false")
	}
	if o.IgnoreVideoEncoderConstraints {
		args = append(args, "ignore_video_encoder_constraints=true")
	}
	if !o.SendFrameMeta {
		args = append(args, "send_frame_meta=false")
	}
	if o.SCID >= 0 {
		args = append(args, fmt.Sprintf("scid=%x", o.SCID))
	}

	// 查询参数
	if o.ListEncoders {
		args = append(args, "list_encoders=true")
	}
	if o.ListDisplays {
		args = append(args, "list_displays=true")
	}
	if o.ListCameras {
		args = append(args, "list_cameras=true")
	}
	if o.ListCameraSizes {
		args = append(args, "list_camera_sizes=true")
	}
	if o.ListApps {
		args = append(args, "list_apps=true")
	}

	return args
}

// formatFloat 格式化浮点数为字符串
func formatFloat(f float64) string {
	// 移除末尾多余的零
	s := strconv.FormatFloat(f, 'f', -1, 64)
	return strings.TrimRight(s, "0")
}
