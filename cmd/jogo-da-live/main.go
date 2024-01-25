package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"

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
	img           *ebiten.Image
	imgFrameCount = 5
	imgSize       = 128
)

var yt *youtube.Client

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("assets/img/gopher_standing.png")
	if err != nil {
		panic(err)
	}
	flag.StringVar(&youtube.LiveId, "y", "Youtube video ID of the stream", "")
}

type Inscrito struct {
	Nome string
	PosX float64
	PosY float64

	Frame int
}

type Jogo struct {
	Inscritos map[string]Inscrito
	Count     int
}

func (j *Jogo) Update() error {
	msg := yt.FetchMessages()
	for _, m := range msg {
		log.Printf("nova mensagem: %#v", m.AuthorDetails.DisplayName)
		inscrito := Inscrito{
			Nome: m.AuthorDetails.DisplayName,
			PosX: float64(imgSize * len(j.Inscritos)),
			PosY: float64(Altura) - float64(imgSize),
		}
		j.Inscritos[m.AuthorDetails.ChannelId] = inscrito
	}

	j.Count++
	return nil
}

func (j *Jogo) Draw(tela *ebiten.Image) {
	tela.Fill(CorVerde)
	for _, i := range j.Inscritos {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(i.PosX, i.PosY)
		frameIdx := (j.Count / 5) % imgFrameCount
		fx, fy := 0, frameIdx*imgSize
		frame := img.SubImage(image.Rect(fx, fy, fx+imgSize, fy+imgSize)).(*ebiten.Image)
		tela.DrawImage(frame, opts)
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
		log.Printf("jogo-da-live: não foi possível iniciar conexão com Youtube: %v", err)
	}

	ebiten.SetWindowSize(Largura, Altura)
	ebiten.SetWindowTitle("Jogo da Live!")
	fmt.Println("Jogo da Live iniciando ...")

	if err := ebiten.RunGame(New()); err != nil {
		panic(err)
	}
}
