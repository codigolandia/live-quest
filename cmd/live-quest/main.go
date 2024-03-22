package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/codigolandia/live-quest/assets"
	"github.com/codigolandia/live-quest/log"
	"github.com/codigolandia/live-quest/message"
	"github.com/codigolandia/live-quest/twitch"
	"github.com/codigolandia/live-quest/youtube"
	"github.com/hajimehoshi/bitmapfont/v3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"

	_ "image/png"
)

var (
	Width  = 1920
	Height = 1080

	// Delay para realizar o auto-save em game ticks
	AutoSaveDelay = 60 * 10

	ColorGreen      = color.RGBA{0, 0xff, 0, 0xff}
	ColorGopherBlue = color.RGBA{0x9c, 0xed, 0xff, 0xff}

	Port string
)

var (
	gopherFrameCount = 5
	gopherSize       = 128

	face font.Face

	gravity = 0.1
)

var (
	yt *youtube.Client
	tw *twitch.Client
)

func init() {
	face = bitmapfont.Face

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
	return path.Join(os.TempDir(), "live-quest.json")
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
		v := g.Viewers[uid]
		gi := &GopherImage{
			img: assets.GopherStanding,
			clr: v.SpriteColor,
		}
		v.Sprite = ebiten.NewImageFromImage(gi)
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
			v = NewViewer()
			v.Name = m.Author
			v.Platform = m.Platform
			v.PosY = float64(Height) - float64(gopherSize)
			v.PosX = float64(rand.Int() * gopherSize)

			g.Viewers[m.UID] = v
			g.UIDs = append(g.UIDs, m.UID)
		}

		g.ParseCommands(m, v)
		v.XP += 10
	}
}

func (g *Game) ParseCommands(m message.Message, v *Viewer) {
	// Processa os comandos
	switch {
	case strings.Contains(m.Text, "!jump"):
		log.D("%s is jumping!", m.UID)
		v.VelY = -100
	case strings.Contains(m.Text, "!color"):
		log.D("%s is changing the Gopher color!", m.UID)
		v.SpriteColor = RandomColor()
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
	for _, uid := range g.UIDs {
		e := g.Viewers[uid]
		e.Draw(screen)
	}
	g.DrawLeaderBoard(screen)
}

func (g *Game) DrawLeaderBoard(screen *ebiten.Image) {
	viewers := make([]*Viewer, 0, len(g.UIDs))
	for _, uid := range g.UIDs {
		viewers = append(viewers, g.Viewers[uid])
	}
	sort.Sort(ByXP(viewers))

	px, py := float64(Width-380), 12.0
	topFive := viewers
	if len(viewers) > 5 {
		topFive = viewers[:5]
	}
	for _, v := range topFive {
		txt := fmt.Sprintf("%v [%02d] %04d XP\n", v.Name, v.Level(), v.XP)
		txtLen := len(txt) * pixelPerChar
		py += 24
		px = float64(Width) - float64(txtLen)
		DrawTextAt(screen, txt, px, py)
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

// TODO(ronoaldo): move to own file, like draw.go
var (
	pixelPerChar           = 2 * 6
	textScaleX, textScaleY = 2.0, 1.8
)

func DrawTextAt(screen *ebiten.Image, txt string, px, py float64) {
	textOpts := &ebiten.DrawImageOptions{}
	textOpts.GeoM.Scale(textScaleX, textScaleY)
	textOpts.GeoM.Translate(px, py)

	textOpts.GeoM.Translate(-2.0, -2.0)
	textOpts.ColorScale.SetR(0.0)
	textOpts.ColorScale.SetG(0.0)
	textOpts.ColorScale.SetB(0.0)
	text.DrawWithOptions(screen, txt, face, textOpts)

	textOpts.GeoM.Translate(3.0, 3.0)
	text.DrawWithOptions(screen, txt, face, textOpts)

	textOpts.GeoM.Translate(-1.0, -1.0)
	textOpts.ColorScale.SetR(1.0)
	textOpts.ColorScale.SetG(1.0)
	textOpts.ColorScale.SetB(1.0)
	text.DrawWithOptions(screen, txt, face, textOpts)
}

func main() {
	flag.Parse()

	g := New()
	g.Autoload()

	var err error
	yt, err = youtube.New(g.YoutubePageToken)
	if err != nil {
		log.E("live-quest: unable to initialize Youtube client: %v", err)
	}
	tw, err = twitch.New()
	if err != nil {
		log.E("live-quest: unable to initialize Twitch client: %v", err)
	}

	ebiten.SetWindowSize(Width, Height)
	ebiten.SetWindowTitle("LiveQuest")

	log.I("game initialized")

	// Listen on http 8080
	http.HandleFunc("/chat", g.ServeChat)
	go http.ListenAndServe(":"+Port, nil)
	log.I("http chat overlay started on port " + Port)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
