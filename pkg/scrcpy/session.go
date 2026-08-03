package scrcpy

import (
	"context"
	"fmt"
	"sync"

	"github.com/yuan71058/go-scrcpy/pkg/adb"
	"github.com/yuan71058/go-scrcpy/pkg/audio"
	"github.com/yuan71058/go-scrcpy/pkg/control"
	"github.com/yuan71058/go-scrcpy/pkg/protocol"
	"github.com/yuan71058/go-scrcpy/pkg/server"
	"github.com/yuan71058/go-scrcpy/pkg/transport"
	"github.com/yuan71058/go-scrcpy/pkg/types"
	"github.com/yuan71058/go-scrcpy/pkg/video"
)

// DeviceSession 设备会话
// 管理单个设备的连接、流和控制
type DeviceSession struct {
	serial     string                 // 设备序列号
	conn       *transport.Connection  // 三通道连接
	handshake  *protocol.Handshake    // 握手数据
	videoDec   *video.H264Decoder     // 视频解码器
	audioDec   *audio.OpusDecoder     // 音频解码器
	clipboard  *control.Clipboard     // 剪贴板管理器
	filePusher *control.FilePusher    // 文件推送器
	deviceInfo *types.DeviceInfo      // 设备信息

	// 通道
	videoChan  chan *video.DecodedFrame // 视频帧通道
	audioChan  chan *audio.AudioData    // 音频数据通道
	controlChan chan []byte             // 控制消息通道

	mu     sync.Mutex
	closed bool
}

// NewSession 创建设备会话
func NewSession(serial string) *DeviceSession {
	return &DeviceSession{
		serial:      serial,
		videoChan:   make(chan *video.DecodedFrame, 60),
		audioChan:   make(chan *audio.AudioData, 100),
		controlChan: make(chan []byte, 100),
	}
}

// Connect 建立与设备的连接
func (s *DeviceSession) Connect(ctx context.Context, adbClient *adb.Client, opts server.Options, localJAR string) error {
	logInfo("建立设备连接 [%s]", s.serial)

	// 启动 server
	launcher := server.NewLauncher(adbClient)
	connInfo, err := launcher.Start(ctx, s.serial, opts, localJAR)
	if err != nil {
		return fmt.Errorf("启动 server 失败: %w", err)
	}

	// 建立三通道连接
	conn, err := transport.NewConnection(connInfo.VideoPort, connInfo.AudioPort, connInfo.ControlPort)
	if err != nil {
		return fmt.Errorf("建立连接失败: %w", err)
	}
	s.conn = conn

	// 读取握手数据
	videoReader := transport.NewProtocolReader(conn.VideoConn)
	handshake, err := protocol.ReadHandshake(videoReader)
	if err != nil {
		conn.Close()
		return fmt.Errorf("读取握手数据失败: %w", err)
	}
	s.handshake = handshake

	// 创建视频解码器
	s.videoDec = video.NewH264Decoder(60)

	// 创建音频解码器（如果启用了音频）
	if opts.Audio {
		s.audioDec = audio.NewOpusDecoder(100)
	}

	// 创建剪贴板管理器
	s.clipboard = control.NewClipboard(s)

	// 创建文件推送器
	s.filePusher = control.NewFilePusher(s)

	// 获取设备信息
	props, _ := adbClient.GetDeviceProperties(ctx, s.serial)
	s.deviceInfo = &types.DeviceInfo{
		Serial:     s.serial,
		Model:      handshake.GetDeviceName(),
		Properties: props,
	}

	logInfo("设备连接成功 [%s]: %s", s.serial, handshake.GetDeviceName())
	return nil
}

// Start 启动会话
// 开始读取视频帧、音频数据和设备消息
func (s *DeviceSession) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("会话已关闭")
	}
	s.mu.Unlock()

	logInfo("启动会话 [%s]", s.serial)

	// 启动视频读取 goroutine
	go s.readVideoLoop(ctx)

	// 启动音频读取 goroutine（如果有音频解码器）
	if s.audioDec != nil {
		go s.readAudioLoop(ctx)
	}

	// 启动控制消息读取 goroutine
	go s.readControlLoop(ctx)

	return nil
}

// readVideoLoop 视频帧读取循环
func (s *DeviceSession) readVideoLoop(ctx context.Context) {
	defer func() {
		logDebug("视频读取循环退出 [%s]", s.serial)
	}()

	videoReader := transport.NewProtocolReader(s.conn.VideoConn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 读取帧元数据
		header, err := protocol.ReadFrameHeader(videoReader)
		if err != nil {
			logError("读取帧头失败 [%s]: %v", s.serial, err)
			return
		}

		// 读取帧数据
		data, err := videoReader.ReadFull(int(header.Size))
		if err != nil {
			logError("读取帧数据失败 [%s]: %v", s.serial, err)
			return
		}

		// 推送到解码器
		if err := s.videoDec.Push(data); err != nil {
			logError("推送帧到解码器失败 [%s]: %v", s.serial, err)
			continue
		}

		// 发送到通道
		frame := &video.DecodedFrame{
			PTS:    header.GetPTS(),
			Data:   data,
			Width:  0, // 从解码器获取
			Height: 0,
			Config: header.IsConfigPacket(),
			Key:    header.IsKeyFrame(),
		}

		select {
		case s.videoChan <- frame:
		default:
			// 通道满，丢弃
			logDebug("视频通道满，丢弃帧 [%s]", s.serial)
		}
	}
}

// readAudioLoop 音频数据读取循环
func (s *DeviceSession) readAudioLoop(ctx context.Context) {
	defer func() {
		logDebug("音频读取循环退出 [%s]", s.serial)
	}()

	audioReader := transport.NewProtocolReader(s.conn.AudioConn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 读取帧元数据
		header, err := protocol.ReadFrameHeader(audioReader)
		if err != nil {
			logError("读取音频帧头失败 [%s]: %v", s.serial, err)
			return
		}

		// 读取帧数据
		data, err := audioReader.ReadFull(int(header.Size))
		if err != nil {
			logError("读取音频数据失败 [%s]: %v", s.serial, err)
			return
		}

		// 推送到解码器
		if err := s.audioDec.Push(data); err != nil {
			logError("推送音频到解码器失败 [%s]: %v", s.serial, err)
			continue
		}

		// 发送到通道
		pkt := &audio.AudioData{
			PTS:  header.GetPTS(),
			Data: data,
		}

		select {
		case s.audioChan <- pkt:
		default:
			// 通道满，丢弃
			logDebug("音频通道满，丢弃包 [%s]", s.serial)
		}
	}
}

// readControlLoop 控制消息读取循环
func (s *DeviceSession) readControlLoop(ctx context.Context) {
	defer func() {
		logDebug("控制消息读取循环退出 [%s]", s.serial)
	}()

	controlReader := transport.NewProtocolReader(s.conn.ControlConn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 读取设备消息
		msg, err := protocol.ReadDeviceMessage(controlReader)
		if err != nil {
			logError("读取设备消息失败 [%s]: %v", s.serial, err)
			return
		}

		logDebug("收到设备消息 [%s]: %s", s.serial, protocol.DeviceMessageToString(msg))

		// 分发消息到相应的处理器
		s.dispatchDeviceMessage(msg)
	}
}

// dispatchDeviceMessage 分发设备消息到相应的处理器
func (s *DeviceSession) dispatchDeviceMessage(msg *protocol.DeviceMessage) {
	switch msg.Type {
	case protocol.DeviceMsgClipboard:
		// 剪贴板消息
		if s.clipboard != nil {
			s.clipboard.HandleMessage(msg.Type, msg.Payload)
		}
	case protocol.DeviceMsgAckClipboard:
		// 剪贴板确认
		if s.clipboard != nil {
			s.clipboard.HandleMessage(msg.Type, msg.Payload)
		}
	case protocol.DeviceMsgUHIDOutput:
		// UHID 输出（预留）
		logDebug("收到 UHID 输出 [%s]", s.serial)
	default:
		// 未知消息类型
		logDebug("未知设备消息类型 [%s]: %d", s.serial, msg.Type)
	}
}

// VideoStream 返回视频帧通道
func (s *DeviceSession) VideoStream() <-chan *video.DecodedFrame {
	return s.videoChan
}

// AudioStream 返回音频数据通道
func (s *DeviceSession) AudioStream() <-chan *audio.AudioData {
	return s.audioChan
}

// GetClipboard 获取剪贴板管理器
func (s *DeviceSession) GetClipboard() *control.Clipboard {
	return s.clipboard
}

// GetFilePusher 获取文件推送器
func (s *DeviceSession) GetFilePusher() *control.FilePusher {
	return s.filePusher
}

// SendControl 发送控制消息
func (s *DeviceSession) SendControl(msg []byte) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("会话已关闭")
	}
	s.mu.Unlock()

	if s.conn == nil || s.conn.ControlConn == nil {
		return fmt.Errorf("控制连接未建立")
	}

	writer := transport.NewProtocolWriter(s.conn.ControlConn)
	return writer.Write(msg)
}

// GetDeviceInfo 获取设备信息
func (s *DeviceSession) GetDeviceInfo() *types.DeviceInfo {
	return s.deviceInfo
}

// GetHandshake 获取握手数据
func (s *DeviceSession) GetHandshake() *protocol.Handshake {
	return s.handshake
}

// GetSerial 获取设备序列号
func (s *DeviceSession) GetSerial() string {
	return s.serial
}

// Close 关闭会话
func (s *DeviceSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	logInfo("关闭会话 [%s]", s.serial)

	// 关闭视频解码器
	if s.videoDec != nil {
		s.videoDec.Close()
	}

	// 关闭音频解码器
	if s.audioDec != nil {
		s.audioDec.Close()
	}

	// 关闭剪贴板管理器
	if s.clipboard != nil {
		s.clipboard.Close()
	}

	// 关闭文件推送器
	if s.filePusher != nil {
		s.filePusher.Close()
	}

	// 关闭连接
	if s.conn != nil {
		s.conn.Close()
	}

	// 关闭通道
	close(s.videoChan)
	close(s.audioChan)
	close(s.controlChan)

	return nil
}

// IsClosed 检查会话是否已关闭
func (s *DeviceSession) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}
