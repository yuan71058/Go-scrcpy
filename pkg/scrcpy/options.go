package scrcpy

import (
	"github.com/yuan71058/go-scrcpy/pkg/server"
)

// Options 启动选项
// 封装 server.Options，提供更简洁的 API
type Options struct {
	// ADB 可执行文件路径，默认 "adb"
	ADBPath string

	// 本地 scrcpy-server.jar 路径（为空则使用设备上已有的）
	LocalJAR string

	// Server 启动参数
	Server server.Options
}

// DefaultOptions 返回默认选项
func DefaultOptions() Options {
	return Options{
		ADBPath: "adb",
		Server: server.Options{
			Video:            true,
			Audio:            true,
			Control:          true,
			VideoBitRate:     8000000,
			AudioBitRate:     128000,
			ClipboardAutosync: true,
			DownsizeOnError:  true,
			Cleanup:          true,
			PowerOn:          true,
			SendFrameMeta:    true,
			SCID:             -1,
			ScreenOffTimeout: -1,
		},
	}
}

// WithVideoCodec 设置视频编解码器
func (o Options) WithVideoCodec(codec string) Options {
	o.Server.VideoCodec = codec
	return o
}

// WithAudioCodec 设置音频编解码器
func (o Options) WithAudioCodec(codec string) Options {
	o.Server.AudioCodec = codec
	return o
}

// WithMaxFps 设置最大帧率
func (o Options) WithMaxFps(fps float64) Options {
	o.Server.MaxFps = fps
	return o
}

// WithMaxSize 设置最大分辨率
func (o Options) WithMaxSize(size int) Options {
	o.Server.MaxSize = size
	return o
}

// WithBitrate 设置视频码率
func (o Options) WithBitrate(bitrate int) Options {
	o.Server.VideoBitRate = bitrate
	return o
}

// WithAudioEnabled 启用/禁用音频
func (o Options) WithAudioEnabled(enabled bool) Options {
	o.Server.Audio = enabled
	return o
}

// WithControlEnabled 启用/禁用控制
func (o Options) WithControlEnabled(enabled bool) Options {
	o.Server.Control = enabled
	return o
}
