package record

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Screenshot 截图
// 从视频帧生成图片
type Screenshot struct {
	width  int
	height int
}

// NewScreenshot 创建截图器
func NewScreenshot(width, height int) *Screenshot {
	return &Screenshot{
		width:  width,
		height: height,
	}
}

// Capture 截取当前帧
// data: 视频帧数据
// format: 输出格式 ("png", "jpeg")
func (s *Screenshot) Capture(data []byte, format string) ([]byte, error) {
	if format == "" {
		format = "png"
	}

	logInfo("截取屏幕: %dx%d, format=%s", s.width, s.height, format)

	// 创建简单的测试图像
	// 实际应用中应从解码后的视频帧生成
	img := image.NewRGBA(image.Rect(0, 0, s.width, s.height))

	// 填充渐变色作为示例
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			r := uint8(x * 255 / s.width)
			g := uint8(y * 255 / s.height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// 编码为指定格式
	var buf bytes.Buffer
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("PNG 编码失败: %w", err)
		}
	case "jpeg":
		// JPEG 编码需要 image/jpeg 包
		return nil, fmt.Errorf("JPEG 格式暂不支持")
	default:
		return nil, fmt.Errorf("不支持的格式: %s", format)
	}

	logInfo("截图成功: %d bytes", buf.Len())
	return buf.Bytes(), nil
}

// SaveToFile 保存截图到文件
func (s *Screenshot) SaveToFile(data []byte, path string, format string) error {
	if format == "" {
		format = "png"
	}

	logInfo("保存截图到: %s", path)

	// 如果没有数据，生成测试图像
	if data == nil {
		var err error
		data, err = s.Capture(nil, format)
		if err != nil {
			return err
		}
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("保存截图失败: %w", err)
	}

	logInfo("截图已保存: %s (%d bytes)", path, len(data))
	return nil
}

// SetSize 设置截图尺寸
func (s *Screenshot) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// GetSize 获取截图尺寸
func (s *Screenshot) GetSize() (int, int) {
	return s.width, s.height
}
