package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"sync"

	"github.com/codigolandia/live-quest/assets"
	"github.com/codigolandia/live-quest/message"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type Viewer struct {
	Name     string `json:"name"`
	UID      string `json:"uid"`
	Platform string `json:"platform"`

	HP int `json:"hp"`
	XP int `json:"xp"`

	PosX float64 `json:"posx"`
	PosY float64 `json:"posy"`

	VelX float64 `json:"velx"`
	VelY float64 `json:"vely"`

	Animation      string     `json:"animation"`
	AnimationFrame int        `json:"animationFrame"`
	SpriteColor    color.RGBA `json:"spriteColor"`

	CompletedChallenges map[string]struct{} `json:"completedChallenges"`

	mu sync.Mutex
}

func NewViewer() *Viewer {
	v := Viewer{
		HP:                  100,
		XP:                  0,
		SpriteColor:         ColorGopherBlue,
		CompletedChallenges: make(map[string]struct{}),
	}
	return &v
}

const XPPerLevel = 100

func (v *Viewer) IncXP(xpDelta int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.XP += xpDelta
}

func (v *Viewer) Level() int {
	return v.XP / XPPerLevel
}

func (v *Viewer) CurrAnim() *assets.Animation {
	if v.Animation == "" {
		v.Animation = "standing"
	}
	a := assets.Bundles["gopher"][v.Animation]
	return a
}

func (v *Viewer) CurrAnimFrames() int {
	return len(v.CurrAnim().Frames)
}

func (v *Viewer) Stop() {
	v.Animation = "standing"
	v.AnimationFrame = rand.Int() % v.CurrAnimFrames()
}

func (v *Viewer) WalkLeft() {
	v.Animation = "walking_left"
	if v.AnimationFrame == 0 {
		v.AnimationFrame = rand.Int() % v.CurrAnimFrames()
	}
}

func (v *Viewer) WalkRight() {
	v.Animation = "walking_right"
	if v.AnimationFrame == 0 {
		v.AnimationFrame = rand.Int() % v.CurrAnimFrames()
	}
}

func (v *Viewer) Jump() {
	v.VelY = -150
}

func (v *Viewer) Damage(value int) {
	v.HP -= value
}

var baseDamage = 10

func (v *Viewer) Attack(other *Viewer) int {
	if v.XP > other.XP {
		dmg := float64(baseDamage) * (float64(v.XP) / float64(other.XP))
		return min(int(dmg), 25) - rand.Intn(6)
	}
	return baseDamage + rand.Intn(16)
}

func (v *Viewer) UpdateAnimation(g *Game) {
	if g.Count%6 == 0 {
		a := v.CurrAnim()
		v.AnimationFrame++
		if v.AnimationFrame >= len(a.Frames) {
			v.AnimationFrame = 0
		}
	}
}

func (v *Viewer) Update(g *Game) {
	v.UpdateAnimation(g)
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
	frame := v.CurrAnim().Frames[v.AnimationFrame]
	opts := &colorm.DrawImageOptions{}
	opts.GeoM.Translate(v.PosX, v.PosY)
	colorTx := colorm.ColorM{}
	colorTx.ScaleWithColor(v.SpriteColor)
	colorm.DrawImage(screen, frame, colorTx, opts)
	if v.AnimationFrame < len(v.CurrAnim().Skins) {
		colorTx.Reset()
		colorm.DrawImage(screen, v.CurrAnim().Skins[v.AnimationFrame], colorTx, opts)
	}

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

	// Draw Platform Icon
	iconOpts := &ebiten.DrawImageOptions{}
	iconOpts.GeoM.Translate(v.PosX-18, v.PosY-40)
	if v.Platform == message.PlatformYoutube {
		screen.DrawImage(assets.YoutubeIcon, iconOpts)
	} else {
		screen.DrawImage(assets.TwitchIcon, iconOpts)
	}

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

func (v *Viewer) MarkCompleted(challenge string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.CompletedChallenges[challenge] = struct{}{}
}

// ByXP sorts a list of viewers by their XP.
type ByXP []*Viewer

func (v ByXP) Len() int { return len(v) }
func (v ByXP) Less(i, j int) bool {
	if v[i].XP == v[j].XP {
		return v[i].Name >= v[j].Name
	}
	return v[i].XP >= v[j].XP
}
func (v ByXP) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
