package assets

import (
	"embed"
	"image"

	_ "image/png"

	"github.com/codigolandia/live-quest/log"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed img animations web
	Assets embed.FS
)

var (
	BarBG   *ebiten.Image
	HPBarFG *ebiten.Image
	XPBarFG *ebiten.Image

	YoutubeIcon *ebiten.Image
	TwitchIcon  *ebiten.Image
)

func init() {
	BarBG = LoadEbitenImg("img/hp_bar_bg.png")
	HPBarFG = LoadEbitenImg("img/hp_bar_fg.png")
	XPBarFG = LoadEbitenImg("img/xp_bar_fg.png")

	YoutubeIcon = LoadEbitenImg("img/youtube_icon.png")
	TwitchIcon = LoadEbitenImg("img/twitch_icon.png")
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
