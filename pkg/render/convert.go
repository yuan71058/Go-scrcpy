package render

// YUV420PtoRGB YUV420P 转 RGB24
func YUV420PtoRGB(yuv []byte, w, h int) []byte {
	ySize := w * h
	uvW := (w + 1) / 2
	uvH := (h + 1) / 2

	rgb := make([]byte, w*h*3)

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			// Y index
			yIdx := j*w + i

			// U/V index (subsampled 2x2)
			uIdx := ySize + (j/2)*uvW + (i/2)
			vIdx := ySize + uvW*uvH + (j/2)*uvW + (i/2)

			if uIdx >= len(yuv) || vIdx >= len(yuv) || yIdx >= len(yuv) {
				continue
			}

			y := int32(yuv[yIdx]) - 16
			u := int32(yuv[uIdx]) - 128
			v := int32(yuv[vIdx]) - 128

			// BT.601
			r := (298*y + 409*v + 128) >> 8
			g := (298*y - 100*u - 208*v + 128) >> 8
			b := (298*y + 516*u + 128) >> 8

			// Clamp
			if r < 0 { r = 0 } else if r > 255 { r = 255 }
			if g < 0 { g = 0 } else if g > 255 { g = 255 }
			if b < 0 { b = 0 } else if b > 255 { b = 255 }

			off := (j*w + i) * 3
			rgb[off] = byte(r)
			rgb[off+1] = byte(g)
			rgb[off+2] = byte(b)
		}
	}
	return rgb
}
