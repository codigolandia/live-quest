package main

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Viewer struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`

	PosX float64 `json:"posx"`
	PosY float64 `json:"posy"`

	VelX float64 `json:"velx"`
	VelY float64 `json:"vely"`

	Sprite *ebiten.Image `json:"-"`

	SpriteColor color.Color `json:"spriteColor"`
	SpriteFrame int         `json:"spriteFrame"`
}

func (v *Viewer) Update(g *Game) {
	if g.Count%12 == 0 {
		v.SpriteFrame++
		if v.SpriteFrame >= imgFrameCount {
			v.SpriteFrame = 0
		}
	}

	if g.Count%128 == 0 {
		switch rand.Intn(3) {
		case 0:
			v.VelX = 0
		case 1:
			v.VelX = 1
		case 2:
			v.VelX = -1
		}
	}

	// Moving
	v.PosX = v.PosX + v.VelX
	v.PosY = v.PosY + v.VelY
	v.VelY = v.VelY + gravity
	if v.VelY < 0 {
		v.VelY = 0
	}

	// Screen border
	screenMargin := 400.0
	if v.PosX > float64(Width)-float64(imgSize)-screenMargin {
		v.PosX = float64(Width) - float64(imgSize) - screenMargin
		v.VelX = -2
	}
	if v.PosX < 0 {
		v.PosX = 0
		v.VelX = 2
	}
	if v.PosY < 0 {
		v.PosY = 0
		v.VelY = 50
	}
	floor := float64(Height - imgSize)
	if v.PosY > floor {
		v.PosY = floor
		v.VelY = 0
	}
}

func (v *Viewer) Draw(screen *ebiten.Image) {
	fx, fy := 0, v.SpriteFrame*imgSize

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(v.PosX, v.PosY)
	frameClip := image.Rect(fx, fy, fx+imgSize, fy+imgSize)
	frame := v.Sprite.SubImage(frameClip).(*ebiten.Image)
	screen.DrawImage(frame, opts)
	text.Draw(screen, v.Name, face, int(v.PosX), int(v.PosY), color.Black)
	text.Draw(screen, v.Name, face, int(v.PosX)+1, int(v.PosY)+1, color.White)
}
