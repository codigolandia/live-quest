package main

import (
	"flag"
	"fmt"
	"github.com/codigolandia/jogo-da-live/youtube"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	"log"
)

var (
	Largura = 1920
	Altura  = 1080

	CorVerde = color.RGBA{0, 0xff, 0, 0xff}
)

var img *ebiten.Image

var yt *youtube.Client

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("assets/img/gopher.png")
	if err != nil {
		panic(err)
	}
	flag.StringVar(&youtube.LiveId, "y", "Youtube video ID of the stream", "")
}

type Inscrito struct {
	Nome string
	PosX float64
	PosY float64
}

type Jogo struct {
	Inscritos []Inscrito
}

func (j *Jogo) Update() error {
	msg := yt.FetchMessages()
	for _, m := range msg {
		log.Printf("novo inscrito: %#v", m)
		j.Inscritos = append(j.Inscritos, Inscrito{
			Nome: m.AuthorDetails.DisplayName,
			PosX: float64(50.0 * len(j.Inscritos)),
			PosY: float64(Altura) - 90,
		})
	}

	if len(j.Inscritos) == 0 {
		j.Inscritos = append(j.Inscritos, Inscrito{
			Nome: "Gopher",
			PosX: 50.0,
			PosY: float64(Altura) - 90,
		})
	}
	return nil
}

func (j *Jogo) Draw(tela *ebiten.Image) {
	tela.Fill(CorVerde)
	for _, i := range j.Inscritos {
		geom := ebiten.GeoM{}
		geom.Scale(0.5, 0.5)
		geom.Translate(i.PosX, i.PosY)
		tela.DrawImage(img, &ebiten.DrawImageOptions{
			GeoM: geom,
		})
	}
}

func (j *Jogo) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return Largura, Altura
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

	if err := ebiten.RunGame(&Jogo{}); err != nil {
		panic(err)
	}
}
