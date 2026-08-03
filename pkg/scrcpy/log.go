// Package scrcpy 提供 Go-scrcpy 的核心 API
// 支持单设备和多设备同时接入控制
package scrcpy

import (
	"fmt"
)

// 日志级别常量
const (
	LogLevelNone = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

var logLevel = LogLevelError

// SetLogLevel 设置 scrcpy 模块的日志级别
func SetLogLevel(level int) {
	logLevel = level
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		fmt.Printf("[SCRCPY DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		fmt.Printf("[SCRCPY INFO] "+format+"\n", args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		fmt.Printf("[SCRCPY ERROR] "+format+"\n", args...)
	}
}
