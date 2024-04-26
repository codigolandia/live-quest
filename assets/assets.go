package assets

import (
	"embed"
	"image"
	"image/color"
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
	BarBG = LoadEbitenImg("img/hp_bar_bg.png", false)
	HPBarFG = LoadEbitenImg("img/hp_bar_fg.png", false)
	XPBarFG = LoadEbitenImg("img/xp_bar_fg.png", false)

	YoutubeIcon = LoadEbitenImg("img/youtube_icon.png", false)
	TwitchIcon = LoadEbitenImg("img/twitch_icon.png", false)
}
func LoadImg(path string, asGrayScale bool) image.Image {
	r, _ := Assets.Open(path)
	img, _, err := image.Decode(r)
	if err != nil {
		log.E("image not found: %v", err)
	}

	if asGrayScale {
		gray := image.NewRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				clr := uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
				rgba := color.RGBA{R: clr, G: clr, B: clr, A: uint8(a)}
				gray.Set(x, y, rgba)
			}
		}
		return gray
	}

	return img
}

func LoadEbitenImg(path string, asGrayScale bool) *ebiten.Image {
	return ebiten.NewImageFromImage(LoadImg(path, asGrayScale))
}
