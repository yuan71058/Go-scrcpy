package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/yuan71058/go-scrcpy/pkg/transport"
)

// DeviceMessage 设备消息结构
// 从 scrcpy-server 接收的设备消息
type DeviceMessage struct {
	Type    byte   // 消息类型
	Payload []byte // 消息载荷
}

// ClipboardMessage 剪贴板消息
type ClipboardMessage struct {
	Text string // 剪贴板文本内容
}

// AckClipboardMessage 剪贴板确认消息
type AckClipboardMessage struct {
	Sequence uint64 // 确认的序列号
}

// UHIDOutputMessage UHID 输出消息
type UHIDOutputMessage struct {
	ID   uint16 // UHID 设备 ID
	Data []byte // 输出数据
}

// PushResponseMessage 文件推送响应消息 (ws-scrcpy 扩展)
type PushResponseMessage struct {
	ID   int16  // 推送 ID
	Code byte   // 响应码
}

// ReadDeviceMessage 从控制通道读取设备消息
func ReadDeviceMessage(reader *transport.ProtocolReader) (*DeviceMessage, error) {
	logDebug("读取设备消息...")

	// 读取消息类型
	msgType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("读取消息类型失败: %w", err)
	}

	var payload []byte

	switch msgType {
	case DeviceMsgClipboard:
		// 剪贴板消息: [type:1][length:4][text:N]
		length, err := reader.ReadUint32BE()
		if err != nil {
			return nil, fmt.Errorf("读取剪贴板长度失败: %w", err)
		}
		text, err := reader.ReadString(int(length))
		if err != nil {
			return nil, fmt.Errorf("读取剪贴板内容失败: %w", err)
		}
		payload = []byte(text)

	case DeviceMsgAckClipboard:
		// 剪贴板确认: [type:1][sequence:8]
		sequence, err := reader.ReadUint64BE()
		if err != nil {
			return nil, fmt.Errorf("读取序列号失败: %w", err)
		}
		payload = make([]byte, 8)
		binary.BigEndian.PutUint64(payload, sequence)

	case DeviceMsgUHIDOutput:
		// UHID 输出: [type:1][id:2][size:2][data:N]
		id, err := reader.ReadUint16BE()
		if err != nil {
			return nil, fmt.Errorf("读取 UHID ID 失败: %w", err)
		}
		size, err := reader.ReadUint16BE()
		if err != nil {
			return nil, fmt.Errorf("读取 UHID 数据大小失败: %w", err)
		}
		data, err := reader.ReadFull(int(size))
		if err != nil {
			return nil, fmt.Errorf("读取 UHID 数据失败: %w", err)
		}
		payload = make([]byte, 4+len(data))
		binary.BigEndian.PutUint16(payload[0:2], id)
		binary.BigEndian.PutUint16(payload[2:4], size)
		copy(payload[4:], data)

	default:
		// 未知消息类型，尝试读取剩余数据
		logDebug("未知设备消息类型: %d", msgType)
		return &DeviceMessage{
			Type: msgType,
		}, nil
	}

	logDebug("设备消息: type=%d, payload=%d bytes", msgType, len(payload))

	return &DeviceMessage{
		Type:    msgType,
		Payload: payload,
	}, nil
}

// ParseClipboardMessage 解析剪贴板消息
func ParseClipboardMessage(msg *DeviceMessage) (*ClipboardMessage, error) {
	if msg.Type != DeviceMsgClipboard {
		return nil, fmt.Errorf("消息类型不是剪贴板: %d", msg.Type)
	}

	return &ClipboardMessage{
		Text: string(msg.Payload),
	}, nil
}

// ParseAckClipboardMessage 解析剪贴板确认消息
func ParseAckClipboardMessage(msg *DeviceMessage) (*AckClipboardMessage, error) {
	if msg.Type != DeviceMsgAckClipboard {
		return nil, fmt.Errorf("消息类型不是剪贴板确认: %d", msg.Type)
	}

	if len(msg.Payload) < 8 {
		return nil, fmt.Errorf("剪贴板确认消息数据不足: %d bytes", len(msg.Payload))
	}

	sequence := binary.BigEndian.Uint64(msg.Payload[:8])
	return &AckClipboardMessage{
		Sequence: sequence,
	}, nil
}

// ParseUHIDOutputMessage 解析 UHID 输出消息
func ParseUHIDOutputMessage(msg *DeviceMessage) (*UHIDOutputMessage, error) {
	if msg.Type != DeviceMsgUHIDOutput {
		return nil, fmt.Errorf("消息类型不是 UHID 输出: %d", msg.Type)
	}

	if len(msg.Payload) < 4 {
		return nil, fmt.Errorf("UHID 输出消息数据不足: %d bytes", len(msg.Payload))
	}

	id := binary.BigEndian.Uint16(msg.Payload[0:2])
	size := binary.BigEndian.Uint16(msg.Payload[2:4])
	data := msg.Payload[4 : 4+int(size)]

	return &UHIDOutputMessage{
		ID:   id,
		Data: data,
	}, nil
}

// ParsePushResponseMessage 解析文件推送响应消息 (ws-scrcpy 扩展)
func ParsePushResponseMessage(msg *DeviceMessage) (*PushResponseMessage, error) {
	if msg.Type != 101 {
		return nil, fmt.Errorf("消息类型不是推送响应: %d", msg.Type)
	}

	if len(msg.Payload) < 3 {
		return nil, fmt.Errorf("推送响应消息数据不足: %d bytes", len(msg.Payload))
	}

	id := int16(binary.BigEndian.Uint16(msg.Payload[0:2]))
	code := msg.Payload[2]

	return &PushResponseMessage{
		ID:   id,
		Code: code,
	}, nil
}

// DeviceMessageToString 设备消息转字符串（用于调试）
func DeviceMessageToString(msg *DeviceMessage) string {
	if msg == nil {
		return "NIL"
	}

	switch msg.Type {
	case DeviceMsgClipboard:
		text := string(msg.Payload)
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		return fmt.Sprintf("CLIPBOARD(text=%q)", text)
	case DeviceMsgAckClipboard:
		if len(msg.Payload) >= 8 {
			sequence := binary.BigEndian.Uint64(msg.Payload[:8])
			return fmt.Sprintf("ACK_CLIPBOARD(sequence=%d)", sequence)
		}
		return "ACK_CLIPBOARD(invalid)"
	case DeviceMsgUHIDOutput:
		return fmt.Sprintf("UHID_OUTPUT(size=%d)", len(msg.Payload))
	default:
		return fmt.Sprintf("UNKNOWN(type=%d, size=%d)", msg.Type, len(msg.Payload))
	}
}

// ReadDeviceMessageFromReader 从 io.Reader 读取设备消息
// 自动创建 ProtocolReader
func ReadDeviceMessageFromReader(conn io.Reader) (*DeviceMessage, error) {
	reader := transport.NewProtocolReader(conn)
	return ReadDeviceMessage(reader)
}

// ReadDeviceMessages 批量读取设备消息
// 返回消息列表，直到遇到错误或超时
func ReadDeviceMessages(conn io.Reader, count int) ([]*DeviceMessage, error) {
	reader := transport.NewProtocolReader(conn)
	var messages []*DeviceMessage

	for i := 0; i < count; i++ {
		msg, err := ReadDeviceMessage(reader)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return messages, fmt.Errorf("读取第 %d 条消息失败: %w", i+1, err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ValidateClipboardText 验证剪贴板文本是否有效
func ValidateClipboardText(text string) error {
	if len(text) > MaxClipboardTextSize {
		return fmt.Errorf("剪贴板文本过长: %d > %d", len(text), MaxClipboardTextSize)
	}
	return nil
}

// ValidateText 注入文本是否有效
func ValidateText(text string) error {
	if len(text) > MaxTextLength {
		return fmt.Errorf("注入文本过长: %d > %d", len(text), MaxTextLength)
	}
	return nil
}

// FormatDeviceMessages 格式化多条设备消息（用于调试）
func FormatDeviceMessages(msgs []*DeviceMessage) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("共 %d 条设备消息:\n", len(msgs)))
	for i, msg := range msgs {
		sb.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, DeviceMessageToString(msg)))
	}
	return sb.String()
}
