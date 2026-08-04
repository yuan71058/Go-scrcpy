package scrcpy

import (
	"os"
	"path/filepath"

	"github.com/yuan71058/go-scrcpy/pkg/server"
)

// Options 启动选项
// 封装 server.Options，提供更简洁的 API
type Options struct {
	// ADB 可执行文件路径，默认 "adb"
	ADBPath string

	// 本地 scrcpy-server.jar 路径（为空则自动搜索）
	LocalJAR string

	// PC 监听端口（0 = 自动分配）
	ListenPort int

	// Server 启动参数
	Server server.Options
}

// DefaultOptions 返回默认选项
func DefaultOptions() Options {
	return Options{
		ADBPath:    "adb",
		LocalJAR:   findDefaultJAR(),
		ListenPort: 0,
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

// findDefaultJAR 自动搜索 scrcpy-server.jar
func findDefaultJAR() string {
	// 1. 当前目录 ./data/scrcpy-server.jar
	if _, err := os.Stat(filepath.Join(".", "data", "scrcpy-server.jar")); err == nil {
		return filepath.Join(".", "data", "scrcpy-server.jar")
	}

	// 2. 可执行文件所在目录
	if exe, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exe), "data", "scrcpy-server.jar")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 3. 向上遍历查找 go.mod 所在目录
	if dir, err := os.Getwd(); err == nil {
		for i := 0; i < 10; i++ {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				path := filepath.Join(dir, "data", "scrcpy-server.jar")
				if _, err := os.Stat(path); err == nil {
					return path
				}
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return ""
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