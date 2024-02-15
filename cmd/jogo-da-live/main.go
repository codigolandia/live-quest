package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"

	"github.com/codigolandia/jogo-da-live/log"
	"github.com/codigolandia/jogo-da-live/message"
	"github.com/codigolandia/jogo-da-live/twitch"
	"github.com/codigolandia/jogo-da-live/youtube"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"

	_ "image/png"
)

var (
	Largura = 1920
	Altura  = 1080

	// Delay para realizar o auto-save em game ticks
	AutoSaveDelay = 60 * 10

	CorVerde = color.RGBA{0, 0xff, 0, 0xff}
)

var (
	//go:embed assets
	assets embed.FS
)

var (
	img           image.Image
	imgFrameCount = 5
	imgSize       = 128

	face font.Face
)

var (
	yt *youtube.Client
	tw *twitch.Client
)

func init() {
	var err error
	r, _ := assets.Open("assets/img/gopher_standing.png")
	img, _, err = image.Decode(r)
	if err != nil {
		log.E("imagem não encontrada: %v", err)
	}

	tt, err := opentype.Parse(gomono.TTF)
	if err != nil {
		panic(fmt.Sprintf("erro ao carregar a fonte: %v", err))
	}
	face, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		panic(fmt.Sprintf("erro ao inicializar a fonte: %v", err))
	}

	// Opções da linha de comandos
	flag.StringVar(&youtube.LiveId, "youtube-stream", "",
		"Ativa o chat do Youtube no vídeo id informado")
}

type GopherImage struct {
	img image.Image
	clr *color.RGBA
}

func (gopher *GopherImage) Color() color.Color {
	if gopher.clr == nil {
		gopher.clr = &color.RGBA{
			R: uint8(rand.Intn(255)),
			G: uint8(rand.Intn(126)),
			B: uint8(rand.Intn(255)),
			A: 0xff,
		}
	}
	return gopher.clr
}

func (gopher *GopherImage) ColorModel() color.Model {
	return gopher.img.ColorModel()
}

func (gopher *GopherImage) Bounds() image.Rectangle {
	return gopher.img.Bounds()
}

func (gopher *GopherImage) At(x, y int) color.Color {
	original := gopher.img.At(x, y)
	r, g, b, a := original.RGBA()
	if r == 0x9c9c && g == 0xeded && b == 0xffff && a == 0xffff {
		original = gopher.Color()
	}
	return original
}

var gravity = 0.1

type Expectador struct {
	Nome       string
	Plataforma string

	PosX float64
	PosY float64

	VelX float64
	VelY float64

	Sprite      *ebiten.Image `json:"-"`
	SpriteColor color.Color
	SpriteFrame int
}

type Jogo struct {
	Expectadores  map[string]*Expectador `json:"expectadores"`
	UIDs          []string               `json:"-"`
	HistoricoChat []message.Message      `json:"historicoChat"`

	YoutubePageToken string `json:"youtubePageToken"`
	Count            int    `json:"-"`
}

func (j *Jogo) Autosave() {
	if j.Count%(AutoSaveDelay) != 0 {
		return
	}
	log.I("iniciando autosave")
	fileName := "/tmp/jogo-da-live.json"
	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.E("erro ao abrir o arquivo: %v", err)
		return
	}
	defer fd.Close()
	j.YoutubePageToken = yt.NextPageToken()

	if err := json.NewEncoder(fd).Encode(j); err != nil {
		log.E("erro ao serializar: %v", err)
		return
	}
	log.I("jogo salvo")
}

func (j *Jogo) Autoload() {
	fileName := "/tmp/jogo-da-live.json"
	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.E("erro ao abrir o arquivo: %v", err)
		return
	}
	defer fd.Close()

	if err := json.NewDecoder(fd).Decode(j); err != nil {
		log.E("erro ao serializar: %v", err)
		return
	}

	for uid := range j.Expectadores {
		e := j.Expectadores[uid]
		e.Sprite = ebiten.NewImageFromImage(&GopherImage{img: img})
		j.UIDs = append(j.UIDs, uid)
	}

	if yt != nil {
		yt.SetPageToken(j.YoutubePageToken)
	}
	log.I("jogo carregado, %v", j)
}

func (j *Jogo) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		log.I("encerrando ...")
		return ebiten.Termination
	}

	j.Autosave()

	var msg []message.Message

	if yt != nil {
		msg = yt.FetchMessages()
	}
	if tw != nil {
		twMsg := tw.FetchMessages()
		msg = append(msg, twMsg...)
	}

	for _, m := range msg {
		log.D("mensagem recebida de [%v]%v: %#s", m.UID, m.Author, m.Text)
		j.HistoricoChat = append(j.HistoricoChat, m)
		e, ok := j.Expectadores[m.UID]
		if !ok {
			e = &Expectador{
				Nome:       m.Author,
				Plataforma: m.Platform,

				PosY: float64(Altura) - float64(imgSize),
				PosX: float64(rand.Int() * imgSize),

				Sprite:      ebiten.NewImageFromImage(&GopherImage{img: img}),
				SpriteFrame: rand.Int() % imgFrameCount,
			}
			j.Expectadores[m.UID] = e
			j.UIDs = append(j.UIDs, m.UID)
		}
		// Processa os comandos
		switch m.Text {
		case "!jump":
			log.D("%s está pulando!", m.UID)
			e.VelY = -100
		}
	}

	count := 1
	for _, uid := range j.UIDs {
		e := j.Expectadores[uid]

		if j.Count%12 == 0 {
			e.SpriteFrame++
			if e.SpriteFrame >= imgFrameCount {
				e.SpriteFrame = 0
			}
		}

		if j.Count%128 == 0 {
			switch rand.Intn(3) {
			case 0:
				e.VelX = 0
			case 1:
				e.VelX = 1
			case 2:
				e.VelX = -1
			}
		}

		// Movimentação
		e.PosX = e.PosX + e.VelX
		e.PosY = e.PosY + e.VelY
		e.VelY = e.VelY + gravity
		if e.VelY < 0 {
			e.VelY = 0
		}

		// Bordas da tela
		screenMargin := 400.0
		if e.PosX > float64(Largura)-float64(imgSize)-screenMargin {
			e.PosX = float64(Largura) - float64(imgSize) - screenMargin
			e.VelX = -2
		}
		if e.PosX < 0 {
			e.PosX = 0
			e.VelX = 2
		}
		if e.PosY < 0 {
			e.PosY = 0
			e.VelY = 50
		}
		chao := float64(Altura - imgSize)
		if e.PosY > chao {
			e.PosY = chao
			e.VelY = 0
		}

		count = count + 1
	}

	j.Count++
	return nil
}

func (j *Jogo) Draw(tela *ebiten.Image) {
	tela.Fill(CorVerde)
	if img != nil {
		for _, uid := range j.UIDs {
			e := j.Expectadores[uid]
			fx, fy := 0, e.SpriteFrame*imgSize

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(e.PosX, e.PosY)
			frameClip := image.Rect(fx, fy, fx+imgSize, fy+imgSize)
			frame := e.Sprite.SubImage(frameClip).(*ebiten.Image)
			tela.DrawImage(frame, opts)
			text.Draw(tela, e.Nome, face, int(e.PosX), int(e.PosY), color.Black)
			text.Draw(tela, e.Nome, face, int(e.PosX)+1, int(e.PosY)+1, color.White)
		}
	}
}

func (j *Jogo) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return Largura, Altura
}

func New() *Jogo {
	j := Jogo{}
	j.Expectadores = make(map[string]*Expectador)
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
	j := New()
	j.Autoload()

	if err := ebiten.RunGame(j); err != nil {
		panic(err)
	}
}
