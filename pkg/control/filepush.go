package control

import (
	"fmt"
	"io"
	"os"
)

// FilePushCommand 文件推送子命令
type FilePushCommand byte

const (
	// FilePushNew 创建新推送
	FilePushNew FilePushCommand = 0
	// FilePushStart 开始推送
	FilePushStart FilePushCommand = 1
	// FilePushAppend 追加数据
	FilePushAppend FilePushCommand = 2
	// FilePushFinish 完成推送
	FilePushFinish FilePushCommand = 3
)

// FilePushHandler 文件推送处理器
type FilePushHandler struct {
	sender Sender
}

// NewFilePushHandler 创建文件推送处理器
func NewFilePushHandler(sender Sender) *FilePushHandler {
	return &FilePushHandler{
		sender: sender,
	}
}

// PushFile 推送文件到设备
// filename: 设备上的文件名
// data: 文件数据
func (h *FilePushHandler) PushFile(filename string, data io.Reader) error {
	logInfo("推送文件: %s", filename)

	// 读取文件数据
	fileData, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("读取文件数据失败: %w", err)
	}

	// 构建推送消息
	// 格式: [type:102][sub_type:1][filename_len:4][filename:N][data_len:4][data:M]
	msg := make([]byte, 0, 10+len(filename)+len(fileData))
	msg = append(msg, 102) // TYPE_FILE_PUSH
	msg = append(msg, byte(FilePushNew))

	// 文件名长度 + 文件名
	filenameBytes := []byte(filename)
	msg = append(msg, byte(len(filenameBytes)>>24))
	msg = append(msg, byte(len(filenameBytes)>>16))
	msg = append(msg, byte(len(filenameBytes)>>8))
	msg = append(msg, byte(len(filenameBytes)))
	msg = append(msg, filenameBytes...)

	// 文件数据长度 + 文件数据
	msg = append(msg, byte(len(fileData)>>24))
	msg = append(msg, byte(len(fileData)>>16))
	msg = append(msg, byte(len(fileData)>>8))
	msg = append(msg, byte(len(fileData)))
	msg = append(msg, fileData...)

	// 发送消息
	if err := h.sender.SendControl(msg); err != nil {
		return fmt.Errorf("发送推送消息失败: %w", err)
	}

	logInfo("文件推送成功: %s (%d bytes)", filename, len(fileData))
	return nil
}

// PushFileFromPath 从文件路径推送文件
func (h *FilePushHandler) PushFileFromPath(filename string, localPath string) error {
	logInfo("从路径推送文件: %s -> %s", localPath, filename)

	// 打开文件
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	return h.PushFile(filename, file)
}

// InstallAPK 安装 APK 文件
func (h *FilePushHandler) InstallAPK(localPath string) error {
	logInfo("安装 APK: %s", localPath)
	return h.PushFileFromPath("install.apk", localPath)
}
