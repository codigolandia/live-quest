package main

import (
	"embed"
	"flag"
	"image"
	"image/color"

	"github.com/codigolandia/jogo-da-live/log"
	"github.com/codigolandia/jogo-da-live/message"
	"github.com/codigolandia/jogo-da-live/twitch"
	"github.com/codigolandia/jogo-da-live/youtube"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	Largura = 1920
	Altura  = 1080

	CorVerde = color.RGBA{0, 0xff, 0, 0xff}
)

var (
	//go:embed assets
	assets embed.FS
)

var (
	img           *ebiten.Image
	imgFrameCount = 5
	imgSize       = 128
)

var (
	yt *youtube.Client
	tw *twitch.Client
)

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFileSystem(assets, "assets/img/gopher_standing.png")
	if err != nil {
		log.E("imagem não encontrada: %v", err)
	}
	flag.StringVar(&youtube.LiveId, "y", "Youtube video ID of the stream", "")
}

type Inscrito struct {
	Nome       string
	Plataforma string
	PosX       float64
	PosY       float64
}

type Jogo struct {
	Inscritos map[string]Inscrito
	Count     int
}

func (j *Jogo) Update() error {
	var msg []message.Message

	if yt != nil {
		msg = yt.FetchMessages()
	}
	if tw != nil {
		twMsg := tw.FetchMessages()
		msg = append(msg, twMsg...)
	}

	for _, m := range msg {
		inscrito := Inscrito{
			Nome:       m.Author,
			Plataforma: m.Platform,

			PosY: float64(Altura) - float64(imgSize),
		}
		j.Inscritos[m.UID] = inscrito
	}

	count := 0
	for uid := range j.Inscritos {
		i := j.Inscritos[uid]
		i.PosX = float64(count * imgSize)
		j.Inscritos[uid] = i
		count = count + 1
	}

	j.Count++
	return nil
}

func (j *Jogo) Draw(tela *ebiten.Image) {
	tela.Fill(CorVerde)
	if img != nil {
		for _, i := range j.Inscritos {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(i.PosX, i.PosY)
			frameIdx := (j.Count / 5) % imgFrameCount
			fx, fy := 0, frameIdx*imgSize
			frame := img.SubImage(image.Rect(fx, fy, fx+imgSize, fy+imgSize)).(*ebiten.Image)
			tela.DrawImage(frame, opts)
		}
	}
}

func (j *Jogo) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return Largura, Altura
}

func New() *Jogo {
	j := Jogo{}
	j.Inscritos = make(map[string]Inscrito)
	return &j
}

func main() {
	flag.Parse()

	var err error
	yt, err = youtube.New()
	if err != nil {
		log.E("jogo-da-live: não foi possível iniciar conexão com Youtube: %v", err)
	}
	tw, err = twitch.New()
	if err != nil {
		log.E("jogo-da-live: não foi possível iniciar conexão com Twitch: %v", err)
	}

	ebiten.SetWindowSize(Largura, Altura)
	ebiten.SetWindowTitle("Jogo da Live!")
	log.I("Jogo da Live iniciando ...")

	if err := ebiten.RunGame(New()); err != nil {
		panic(err)
	}
}
