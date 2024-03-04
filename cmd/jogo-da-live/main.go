package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"net/http"
	"os"
	"path"

	"github.com/codigolandia/jogo-da-live/assets"
	"github.com/codigolandia/jogo-da-live/log"
	"github.com/codigolandia/jogo-da-live/message"
	"github.com/codigolandia/jogo-da-live/twitch"
	"github.com/codigolandia/jogo-da-live/youtube"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"

	_ "image/png"
)

var (
	Width  = 1920
	Height = 1080

	// Delay para realizar o auto-save em game ticks
	AutoSaveDelay = 60 * 10

	ColorGreen = color.RGBA{0, 0xff, 0, 0xff}

	Port string
)

var (
	img           image.Image
	imgFrameCount = 5
	imgSize       = 128

	face font.Face

	gravity = 0.1
)

var (
	yt *youtube.Client
	tw *twitch.Client
)

func init() {
	var err error
	r, _ := assets.Assets.Open("img/gopher_standing.png")
	img, _, err = image.Decode(r)
	if err != nil {
		log.E("image not found: %v", err)
	}

	tt, err := opentype.Parse(gomono.TTF)
	if err != nil {
		panic(fmt.Sprintf("error loading font: %v", err))
	}
	face, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(fmt.Sprintf("error initializing font: %v", err))
	}

	// Command line options
	flag.StringVar(&youtube.LiveId, "youtube-stream", "",
		"Enable Youtube chat for the provided video ID")
	flag.StringVar(&Port, "port", "8080", "HTTP Port to listen to")
}

type Game struct {
	Viewers     map[string]*Viewer `json:"viewers"`
	UIDs        []string           `json:"-"`
	ChatHistory []message.Message  `json:"chatHistory"`

	YoutubePageToken string `json:"youtubePageToken"`
	Count            int    `json:"-"`
}

func (g *Game) tempFile() string {
	return path.Join(os.TempDir(), "jogo-da-live.json")
}

func (g *Game) Autosave() {
	if g.Count%(AutoSaveDelay) != 0 {
		return
	}
	log.I("auto-saving ...")
	fileName := g.tempFile()
	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.E("error opening tempfile: %v", err)
		return
	}
	defer fd.Close()
	g.YoutubePageToken = yt.NextPageToken()
	enc := json.NewEncoder(fd)
	enc.SetIndent("", "  ")

	if err := enc.Encode(g); err != nil {
		log.E("error serializing data: %v", err)
		return
	}
	log.I("game saved!")
}

func (g *Game) Autoload() {
	fileName := g.tempFile()
	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.E("error opening temp file: %v", err)
		return
	}
	defer fd.Close()

	if err := json.NewDecoder(fd).Decode(g); err != nil {
		log.E("error deserializing: %v", err)
		return
	}

	for uid := range g.Viewers {
		e := g.Viewers[uid]
		e.Sprite = ebiten.NewImageFromImage(&GopherImage{img: img})
		g.UIDs = append(g.UIDs, uid)
	}

	log.I("game loaded")
}

func (g *Game) CheckNewMessages() {
	var msg []message.Message
	if yt != nil {
		msg = yt.FetchMessages()
	}
	if tw != nil {
		twMsg := tw.FetchMessages()
		msg = append(msg, twMsg...)
	}

	for _, m := range msg {
		log.D("new message from [%v]%v: %#s", m.UID, m.Author, m.Text)
		g.ChatHistory = append(g.ChatHistory, m)
		v, ok := g.Viewers[m.UID]
		if !ok {
			v = &Viewer{
				Name:     m.Author,
				Platform: m.Platform,

				PosY: float64(Height) - float64(imgSize),
				PosX: float64(rand.Int() * imgSize),

				Sprite:      ebiten.NewImageFromImage(&GopherImage{img: img}),
				SpriteFrame: rand.Int() % imgFrameCount,
			}
			g.Viewers[m.UID] = v
			g.UIDs = append(g.UIDs, m.UID)
		}

		g.ParseCommands(m, v)
	}
}

func (g *Game) ParseCommands(m message.Message, v *Viewer) {
	// Processa os comandos
	switch m.Text {
	case "!jump":
		log.D("%s is jumping!", m.UID)
		v.VelY = -100
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		log.I("closing ...")
		return ebiten.Termination
	}
	g.Autosave()
	g.CheckNewMessages()
	for _, uid := range g.UIDs {
		v := g.Viewers[uid]
		v.Update(g)
	}
	g.Count++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorGreen)
	if img != nil {
		for _, uid := range g.UIDs {
			e := g.Viewers[uid]
			e.Draw(screen)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return Width, Height
}

func New() *Game {
	g := Game{}
	g.Viewers = make(map[string]*Viewer)
	return &g
}

func (g *Game) ServeChat(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, g.tempFile())
}

func main() {
	flag.Parse()

	g := New()
	g.Autoload()

	var err error
	yt, err = youtube.New(g.YoutubePageToken)
	if err != nil {
		log.E("jogo-da-live: unable to initialize Youtube client: %v", err)
	}
	tw, err = twitch.New()
	if err != nil {
		log.E("jogo-da-live: unable to initialize Twitch client: %v", err)
	}

	ebiten.SetWindowSize(Width, Height)
	ebiten.SetWindowTitle("Game da Live!")

	log.I("game initialized")

	// Listen on http 8080
	http.HandleFunc("/chat", g.ServeChat)
	go http.ListenAndServe(":"+Port, nil)
	log.I("http chat overlay started on port " + Port)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
