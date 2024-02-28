package main

import (
	"image"
	"image/color"
	"math/rand"
)

type GopherImage struct {
	img image.Image
	clr *color.RGBA
}

func (gopher *GopherImage) Color() color.Color {
	if gopher.clr == nil {
		gopher.clr = &color.RGBA{
			R: uint8(rand.Intn(255)),
			G: uint8(rand.Intn(126)),
			B: uint8(rand.Intn(255)),
			A: 0xff,
		}
	}
	return gopher.clr
}

func (gopher *GopherImage) ColorModel() color.Model {
	return gopher.img.ColorModel()
}

func (gopher *GopherImage) Bounds() image.Rectangle {
	return gopher.img.Bounds()
}

func (gopher *GopherImage) At(x, y int) color.Color {
	original := gopher.img.At(x, y)
	r, g, b, a := original.RGBA()
	if r == 0x9c9c && g == 0xeded && b == 0xffff && a == 0xffff {
		original = gopher.Color()
	}
	return original
}
