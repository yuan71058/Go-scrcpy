package transport

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ProtocolWriter 协议数据写入器
// 用于向 socket 写入二进制协议数据
type ProtocolWriter struct {
	w io.Writer
}

// NewProtocolWriter 创建协议写入器
func NewProtocolWriter(w io.Writer) *ProtocolWriter {
	return &ProtocolWriter{
		w: w,
	}
}

// Write 写入数据
func (pw *ProtocolWriter) Write(data []byte) error {
	_, err := pw.w.Write(data)
	if err != nil {
		return fmt.Errorf("写入数据失败: %w", err)
	}
	return nil
}

// WriteByte 写入单个字节
func (pw *ProtocolWriter) WriteByte(b byte) error {
	return pw.Write([]byte{b})
}

// WriteUint16BE 写入 big-endian uint16
func (pw *ProtocolWriter) WriteUint16BE(val uint16) error {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, val)
	return pw.Write(data)
}

// WriteUint32BE 写入 big-endian uint32
func (pw *ProtocolWriter) WriteUint32BE(val uint32) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, val)
	return pw.Write(data)
}

// WriteUint64BE 写入 big-endian uint64
func (pw *ProtocolWriter) WriteUint64BE(val uint64) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, val)
	return pw.Write(data)
}

// WriteInt32BE 写入 big-endian int32
func (pw *ProtocolWriter) WriteInt32BE(val int32) error {
	return pw.WriteUint32BE(uint32(val))
}

// WriteInt64BE 写入 big-endian int64
func (pw *ProtocolWriter) WriteInt64BE(val int64) error {
	return pw.WriteUint64BE(uint64(val))
}

// WriteString 写入固定长度字符串
func (pw *ProtocolWriter) WriteString(s string, length int) error {
	data := make([]byte, length)
	copy(data, []byte(s))
	return pw.Write(data)
}

// WriteUTF8String 写入带 4 字节长度前缀的 UTF-8 字符串
func (pw *ProtocolWriter) WriteUTF8String(s string) error {
	if err := pw.WriteUint32BE(uint32(len(s))); err != nil {
		return err
	}
	return pw.Write([]byte(s))
}
