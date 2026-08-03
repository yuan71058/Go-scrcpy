package transport

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ProtocolReader 带缓冲的协议数据读取器
// 用于从 socket 读取二进制协议数据
type ProtocolReader struct {
	r   io.Reader
	buf []byte
}

// NewProtocolReader 创建协议读取器
func NewProtocolReader(r io.Reader) *ProtocolReader {
	return &ProtocolReader{
		r:   r,
		buf: make([]byte, 4096),
	}
}

// ReadFull 读取指定字节数的数据
// 确保读取完整的 n 字节，否则返回错误
func (pr *ProtocolReader) ReadFull(n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := io.ReadFull(pr.r, data)
	if err != nil {
		return nil, fmt.Errorf("读取 %d 字节失败: %w", n, err)
	}
	return data, nil
}

// ReadByte 读取单个字节
func (pr *ProtocolReader) ReadByte() (byte, error) {
	data, err := pr.ReadFull(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// ReadUint16BE 读取 big-endian uint16
func (pr *ProtocolReader) ReadUint16BE() (uint16, error) {
	data, err := pr.ReadFull(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(data), nil
}

// ReadUint32BE 读取 big-endian uint32
func (pr *ProtocolReader) ReadUint32BE() (uint32, error) {
	data, err := pr.ReadFull(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

// ReadUint64BE 读取 big-endian uint64
func (pr *ProtocolReader) ReadUint64BE() (uint64, error) {
	data, err := pr.ReadFull(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(data), nil
}

// ReadInt32BE 读取 big-endian int32
func (pr *ProtocolReader) ReadInt32BE() (int32, error) {
	val, err := pr.ReadUint32BE()
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// ReadInt64BE 读取 big-endian int64
func (pr *ProtocolReader) ReadInt64BE() (int64, error) {
	val, err := pr.ReadUint64BE()
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// ReadInt64Native 读取 native byte order int64
// 帧元数据使用 native byte order
func (pr *ProtocolReader) ReadInt64Native() (int64, error) {
	data, err := pr.ReadFull(8)
	if err != nil {
		return 0, err
	}
	// Android 通常是 little-endian
	val := int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24 |
		int64(data[4])<<32 | int64(data[5])<<40 | int64(data[6])<<48 | int64(data[7])<<56
	return val, nil
}

// ReadInt32Native 读取 native byte order int32
func (pr *ProtocolReader) ReadInt32Native() (int32, error) {
	data, err := pr.ReadFull(4)
	if err != nil {
		return 0, err
	}
	val := int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16 | int32(data[3])<<24
	return val, nil
}

// ReadString 读取指定长度的字符串
func (pr *ProtocolReader) ReadString(n int) (string, error) {
	data, err := pr.ReadFull(n)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadUTF8String 读取 UTF-8 字符串（带长度前缀）
// 先读取 4 字节长度，再读取字符串内容
func (pr *ProtocolReader) ReadUTF8String() (string, error) {
	length, err := pr.ReadUint32BE()
	if err != nil {
		return "", err
	}
	return pr.ReadString(int(length))
}
