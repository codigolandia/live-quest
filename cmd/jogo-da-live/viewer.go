package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"github.com/codigolandia/jogo-da-live/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

type Viewer struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`

	HP int `json:"hp"`
	XP int `json:"xp"`

	PosX float64 `json:"posx"`
	PosY float64 `json:"posy"`

	VelX float64 `json:"velx"`
	VelY float64 `json:"vely"`

	Sprite      *ebiten.Image `json:"-"`
	SpriteColor *color.RGBA   `json:"spriteColor"`

	SpriteFrameCount int `json:"spriteFrameCount"`
	SpriteFrame      int `json:"spriteFrame"`
}

func NewViewer() *Viewer {
	v := Viewer{
		HP:          100,
		XP:          0,
		SpriteColor: &ColorGopherBlue,
	}
	return &v
}

const XPPerLevel = 100

func (v *Viewer) Level() int {
	return v.XP / XPPerLevel
}

func (v *Viewer) Stop() {
	v.Sprite = ebiten.NewImageFromImage(&GopherImage{
		img: assets.GopherStanding,
		clr: v.SpriteColor,
	})
	v.SpriteFrame = rand.Int() % assets.GopherStandingFrames
	v.SpriteFrameCount = assets.GopherStandingFrames
}

func (v *Viewer) WalkLeft() {
	v.Sprite = ebiten.NewImageFromImage(&GopherImage{
		img: assets.GopherWalkingLeft,
		clr: v.SpriteColor,
	})
	v.SpriteFrame = rand.Int() % assets.GopherWalkingLeftFrames
	v.SpriteFrameCount = assets.GopherWalkingLeftFrames
}

func (v *Viewer) WalkRight() {
	v.Sprite = ebiten.NewImageFromImage(&GopherImage{
		img: assets.GopherWalkingRight,
		clr: v.SpriteColor,
	})
	v.SpriteFrame = rand.Int() % assets.GopherWalkingRightFrames
	v.SpriteFrameCount = assets.GopherWalkingRightFrames
}

func (v *Viewer) Update(g *Game) {
	if g.Count%6 == 0 {
		v.SpriteFrame++
		if v.SpriteFrame >= v.SpriteFrameCount {
			v.SpriteFrame = 0
		}
	}

	if g.Count%200 == 0 {
		switch rand.Intn(3) {
		case 0:
			v.VelX = 0
			v.Stop()
		case 1:
			v.VelX = 0.7
			v.WalkRight()
		case 2:
			v.VelX = -0.7
			v.WalkLeft()
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
	if v.PosX > float64(Width)-float64(gopherSize)-screenMargin {
		v.PosX = float64(Width) - float64(gopherSize) - screenMargin
		v.VelX = -2
		v.WalkLeft()
	}
	if v.PosX < 0 {
		v.PosX = 0
		v.VelX = 2
		v.WalkRight()
	}
	if v.PosY < 0 {
		v.PosY = 0
		v.VelY = 50
	}
	floor := float64(Height - gopherSize)
	if v.PosY > floor {
		v.PosY = floor
		v.VelY = 0
	}
}

func (v *Viewer) Draw(screen *ebiten.Image) {
	fx, fy := 0, v.SpriteFrame*gopherSize

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(v.PosX, v.PosY)
	frameClip := image.Rect(fx, fy, fx+gopherSize, fy+gopherSize)
	frame := v.Sprite.SubImage(frameClip).(*ebiten.Image)
	screen.DrawImage(frame, opts)

	// Operations
	hpBarH := float64(8)
	xpBarH := float64(8)

	// Draw HP Bar
	barOpts := &ebiten.DrawImageOptions{}
	barOpts.GeoM.Translate(v.PosX, v.PosY-(textScaleY*12)-hpBarH-1)
	screen.DrawImage(assets.BarBG, barOpts)
	barOpts.GeoM.Reset()
	barOpts.GeoM.Scale(float64(v.HP)/100.0, 1)
	barOpts.GeoM.Translate(v.PosX, v.PosY-(textScaleY*12)-hpBarH)
	screen.DrawImage(assets.HPBarFG, barOpts)

	// Draw XP Bar
	barOpts = &ebiten.DrawImageOptions{}
	barOpts.GeoM.Translate(v.PosX, v.PosY-(textScaleY*12)-hpBarH-xpBarH-1)
	screen.DrawImage(assets.BarBG, barOpts)
	barOpts.GeoM.Reset()
	// 2||........300
	// (xpAtual - xpNivelAtual) / xpPorNivel
	xpProgress := v.XP - (v.Level() * XPPerLevel)
	barOpts.GeoM.Scale(float64(xpProgress)/float64(XPPerLevel), 1)
	barOpts.GeoM.Translate(v.PosX, v.PosY-(textScaleY*12)-hpBarH-xpBarH)
	screen.DrawImage(assets.XPBarFG, barOpts)

	// Draw viewer name with shadow
	nameTag := fmt.Sprintf("%s [%v]", v.Name, v.Level())
	nameTagLen := len(nameTag) * pixelPerChar

	//  |...[..........].
	//  |.....PX...........
	//  |.....|^^^^^|......
	px, py := v.PosX-(float64(nameTagLen)-float64(gopherSize))/2, int(v.PosY)-2
	DrawTextAt(screen, nameTag, float64(px), float64(py))
}

// ByXP sorts a list of viewers by their XP.
type ByXP []*Viewer

func (v ByXP) Len() int           { return len(v) }
func (v ByXP) Less(i, j int) bool { return v[i].XP >= v[j].XP }
func (v ByXP) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
