package assets

import (
	"embed"
	"image"

	"github.com/codigolandia/jogo-da-live/log"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed img
	Assets embed.FS
)

var (
	GopherStanding       image.Image
	GopherStandingFrames = 5

	GopherWalkingLeft       image.Image
	GopherWalkingLeftFrames = 8

	GopherWalkingRight       image.Image
	GopherWalkingRightFrames = 8

	BarBG *ebiten.Image

	HPBarFG *ebiten.Image
	XPBarFG *ebiten.Image
)

func init() {
	GopherStanding = LoadImg("img/gopher_standing.png")
	GopherWalkingLeft = LoadImg("img/gopher_walking_left.png")
	GopherWalkingRight = LoadImg("img/gopher_walking_right.png")
	BarBG = LoadEbitenImg("img/hp_bar_bg.png")
	HPBarFG = LoadEbitenImg("img/hp_bar_fg.png")
	XPBarFG = LoadEbitenImg("img/xp_bar_fg.png")
}
func LoadImg(path string) image.Image {
	r, _ := Assets.Open(path)
	img, _, err := image.Decode(r)
	if err != nil {
		log.E("image not found: %v", err)
	}
	return img
}

func LoadEbitenImg(path string) *ebiten.Image {
	return ebiten.NewImageFromImage(LoadImg(path))
}
