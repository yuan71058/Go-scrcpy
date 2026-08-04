package render

// YUV420PtoRGB YUV420P 转 RGB565
func YUV420PtoRGB(yuv []byte, w, h int) []byte {
	ySize := w * h
	uvW := (w + 1) / 2
	uvH := (h + 1) / 2

	rgb := make([]byte, w*h*2) // RGB565

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			yIdx := j*w + i
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

			if r < 0 { r = 0 } else if r > 255 { r = 255 }
			if g < 0 { g = 0 } else if g > 255 { g = 255 }
			if b < 0 { b = 0 } else if b > 255 { b = 255 }

			// RGB565: RRRRR GGGGGG BBBBB
			val := uint16((int32(r)>>3)<<11 | (int32(g)>>2)<<5 | (int32(b) >> 3))
			off := (j*w + i) * 2
			rgb[off] = byte(val)
			rgb[off+1] = byte(val >> 8)
		}
	}
	return rgb
}