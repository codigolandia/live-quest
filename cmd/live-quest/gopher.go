package main

import (
	"image/color"
	"math/rand"
)

func RandomColor() *color.RGBA {
	return &color.RGBA{
		R: uint8(rand.Intn(255)),
		G: uint8(rand.Intn(126)),
		B: uint8(rand.Intn(255)),
		A: 0xff,
	}
}
