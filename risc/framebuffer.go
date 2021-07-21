// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package risc

import (
	"image"
	"image/color"
)

var (
	colorBlack = color.RGBA{R: 0x65, G: 0x7b, B: 0x83, A: 0xff}
	colorWhite = color.RGBA{R: 0xfd, G: 0xf6, B: 0xe3, A: 0xff}
)

type Framebuffer struct {
	Rect image.Rectangle
	Pix  []uint32
}

func (fb *Framebuffer) ColorModel() color.Model {
	return color.RGBAModel
}

func (fb *Framebuffer) Bounds() image.Rectangle {
	return fb.Rect
}

func (fb *Framebuffer) At(x, y int) color.Color {
	if !(image.Point{X: x, Y: y}.In(fb.Rect)) {
		return color.RGBA{}
	}
	i, bit := fb.PixOffset(x, y)
	if (fb.Pix[i]>>bit)&1 == 0 {
		return colorBlack
	}
	return colorWhite
}

// PixOffset is a helper method to locate a pixel in the Pix slice.
// It returns the index in Pix for the pixel at coordinates (x, y) and the bit
// position of the pixel within the word, i.e. the pixel value can be isolated
// with (fb.Pix[i]>>bit)&1.
//
// The coordinates (x, y) are interpreted as image coordinates with the
// origin (0, 0) at the top left.
func (fb *Framebuffer) PixOffset(x, y int) (i, bit int) {
	return (fb.Rect.Max.Y-y)*(fb.Rect.Max.X/32) + (x / 32), x % 32
}
