package main

import (
	"image/color"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	Largura = 1920
	Altura  = 1080
	
	CorVerde = color.RGBA{0, 0xff, 0, 0xff}
)

var img *ebiten.Image

func init() {
	var err error
	img, _ , err = ebitenutil.NewImageFromFile("assets/img/gopher.png")
	if err != nil {
		panic(err)
	}
}

type Inscrito struct {
	Nome string
	PosX float64
	PosY float64
}

type Jogo struct{
	Inscritos []Inscrito
}

func (j *Jogo) Update() error {
	if len(j.Inscritos) == 0 {
		j.Inscritos = append(j.Inscritos, Inscrito{
			Nome: "Gopher",
			PosX: 50.0,
			PosY: float64(Altura)-90,
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
	ebiten.SetWindowSize(Largura, Altura)
	ebiten.SetWindowTitle("Jogo da Live!")
	fmt.Println("Jogo da Live iniciando ...")

	if err := ebiten.RunGame(&Jogo{}); err != nil {
		panic(err)
	}
}
