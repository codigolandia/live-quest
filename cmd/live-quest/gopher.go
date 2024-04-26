package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
)

func SelectColor(message string) (c color.RGBA) {
	err := fmt.Errorf("no color")
	parts := strings.Fields(message)

	if len(parts) >= 2 {
		switch len(parts[1]) {
		case 7:
			_, err = fmt.Sscanf(parts[1], "#%02x%02x%02x", &c.R, &c.G, &c.B)
		case 4:
			_, err = fmt.Sscanf(parts[1], "#%1x%1x%1x", &c.R, &c.G, &c.B)
			// Double the hex digits:
			c.R *= 17
			c.G *= 17
			c.B *= 17
		default:
			err = fmt.Errorf("invalid length, must be 7 or 4")
		}
	}
	if err != nil {
		c.R = uint8(rand.Intn(255))
		c.G = uint8(rand.Intn(126))
		c.B = uint8(rand.Intn(255))
	}

	c.A = 0xff
	return c
}
