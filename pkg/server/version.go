// Package server 管理 scrcpy-server 的启动、配置和生命周期
package server

// 当前支持的 scrcpy-server 版本
const (
	// ServerVersion scrcpy-server 版本号，必须与设备上的 JAR 匹配
	ServerVersion = "4.1"

	// ServerPath 设备上 scrcpy-server JAR 的存储路径
	ServerPath = "/data/local/tmp/scrcpy-server.jar"

	// PIDFile 设备上 PID 文件的路径
	PIDFile = "/data/local/tmp/go_scrcpy.pid"
)
