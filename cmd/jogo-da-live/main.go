package main

import (
	"embed"
	"encoding/json"
	"flag"
	"image"
	"image/color"
	"os"

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
	// Opções da linha de comandos
	flag.StringVar(&youtube.LiveId, "youtube-stream", "",
		"Ativa o chat do Youtube no vídeo id informado")
}

type Expectador struct {
	Nome       string
	Plataforma string
	PosX       float64
	PosY       float64
}

type Jogo struct {
	Expectadores  map[string]Expectador `json:"expectadores"`
	HistoricoChat []message.Message     `json:"historicoChat"`

	YoutubePageToken string `json:"youtubePageToken"`
	Count            int    `json:"-"`
}

func (j *Jogo) Autosave() {
	if j.Count%(60*10) != 0 {
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
		inscrito := Expectador{
			Nome:       m.Author,
			Plataforma: m.Platform,

			PosY: float64(Altura) - float64(imgSize),
		}
		j.Expectadores[m.UID] = inscrito
	}

	count := 1
	for uid := range j.Expectadores {
		i := j.Expectadores[uid]
		i.PosX = float64(count * imgSize)
		log.D("posicionando %v: %v,%v", uid, i.PosX, i.PosY)
		j.Expectadores[uid] = i
		count = count + 1
	}

	j.Count++
	return nil
}

func (j *Jogo) Draw(tela *ebiten.Image) {
	tela.Fill(CorVerde)
	if img != nil {
		for _, i := range j.Expectadores {
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
	j.Expectadores = make(map[string]Expectador)
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
