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
	"sync"

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

	Port          string
	HttpHotReload bool
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
	flag.StringVar(&Port, "http-port", "8080", "HTTP Port to listen to")
	flag.BoolVar(&HttpHotReload, "http-hot-reload", false, "Hot-reload of http assets when developing")
}

type FightState struct {
	Player1     string `json:"player1"`
	Player2     string `json:"player2"`
	CurrentTurn string `json:"currentTurn"`
}

type Game struct {
	Viewers     map[string]*Viewer `json:"viewers"`
	UIDs        []string           `json:"-"`
	ChatHistory []message.Message  `json:"chatHistory"`

	FightingQueue map[string]bool `json:"fightingQueue"`

	FightState FightState `json:"fightState"`

	YoutubePageToken string `json:"youtubePageToken"`
	Count            int    `json:"-"`
}

var tempFileMu sync.Mutex

func (g *Game) tempFile() string {
	return path.Join(os.TempDir(), "live-quest.json")
}

func (g *Game) Autosave() {
	tempFileMu.Lock()
	defer tempFileMu.Unlock()

	if g.Count%(AutoSaveDelay) != 0 {
		return
	}
	log.I("auto-saving ...")
	fileName := g.tempFile()
	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
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
		v.UID = uid
		v.HP = 100
		g.UIDs = append(g.UIDs, uid)
	}

	// Sanity Check
	if g.FightingQueue == nil {
		g.FightingQueue = make(map[string]bool)
	}
	if g.Viewers == nil {
		g.Viewers = make(map[string]*Viewer)
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
			v.UID = m.UID
			v.PosY = float64(Height) / 2
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
		log.D("%s is jumping!", m.Author)
		v.VelY = -100
	case strings.Contains(m.Text, "!color"):
		log.D("%s is changing the Gopher color!", m.Author)
		v.SpriteColor = RandomColor()
	case strings.Contains(m.Text, "!fight"):
		log.D("%s is looking for a fight!", m.Author)
		g.FightingQueue[m.UID] = true
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
		if g.FightState.CurrentTurn == "" {
			v.Update(g)
		} else {
			v.UpdateAnimation(g)
		}
	}
	g.FightRound()
	g.Count++
	return nil
}

func (g *Game) UpdateFightPositions() {
	p1 := g.Viewers[g.FightState.Player1]
	p1.WalkRight()
	p1.VelX = 0
	p1.PosX = 1000

	p2 := g.Viewers[g.FightState.Player2]
	p2.WalkLeft()
	p2.VelX = 0
	p2.PosX = 1200

	crowd := make([]*Viewer, 0, len(g.Viewers))
	for uid, _ := range g.Viewers {
		if uid == p1.UID || uid == p2.UID {
			continue
		}
		v := g.Viewers[uid]
		crowd = append(crowd, v)
	}
	sort.Sort(ByXP(crowd))

	dx := 800/len(g.Viewers) + 1
	px := 64
	for i := range crowd {
		v := crowd[i]
		v.PosX = float64(px)
		v.WalkRight()
		v.VelX = 0
		px = px + dx
	}
}

func (g *Game) RemoveFromQueue(p1, p2 string) {
	delete(g.FightingQueue, p1)
	delete(g.FightingQueue, p2)
}

func (g *Game) FightRound() {
	if len(g.FightingQueue) < 2 && g.FightState.CurrentTurn == "" {
		return
	}
	if g.Count%60 != 0 {
		return
	}
	if g.FightState.CurrentTurn == "" {
		log.D("fight: initializing fight ...")
		fighters := g.SortFighters()
		g.FightState.Player1 = fighters[0].UID
		g.FightState.Player2 = fighters[1].UID
		g.FightState.CurrentTurn = g.FightState.Player2
		g.RemoveFromQueue(fighters[0].UID, fighters[1].UID)
		g.UpdateFightPositions()
	}

	log.D("fight: %s vs %s...", g.FightState.Player1, g.FightState.Player2)
	attacker := g.Viewers[g.FightState.CurrentTurn]
	defender := g.Viewers[g.FightState.Player1]
	if g.FightState.CurrentTurn == g.FightState.Player1 {
		defender = g.Viewers[g.FightState.Player2]
	}

	dmg := attacker.Attack(defender)
	log.D("fight: %s damaged %s by %d", attacker.Name, defender.Name, dmg)
	defender.Damage(dmg)
	g.FightState.CurrentTurn = defender.UID

	if defender.HP <= 0 {
		log.D("fight: fight is over: %v won!", attacker.Name)
		g.FightState = FightState{}
		defender.HP = 100
		defender.XP += 20
		attacker.HP = 100
		attacker.XP += 50
		attacker.VelY = -150
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorGreen)
	for _, uid := range g.UIDs {
		v := g.Viewers[uid]
		v.Draw(screen)
	}
	g.DrawLeaderBoard(screen)
}

func (g *Game) DrawLeaderBoard(screen *ebiten.Image) {
	// Top 5
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
	// Fighting queue
	py += 24
	px = float64(Width) - 350
	if len(g.FightingQueue) > 0 {
		fighters := g.SortFighters()
		py += 24
		DrawTextAt(screen, "** Fighting Queue **", px, py)
		py += 24
		for _, v := range fighters {
			DrawTextAt(screen, v.Name, px, py)
			py += 24
		}
	}
}

func (g *Game) SortFighters() []*Viewer {
	fighters := make([]*Viewer, 0, len(g.FightingQueue))
	for uid, _ := range g.FightingQueue {
		v := g.Viewers[uid]
		fighters = append(fighters, v)
	}
	sort.Sort(ByXP(fighters))
	return fighters
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return Width, Height
}

func New() *Game {
	g := Game{}
	g.Viewers = make(map[string]*Viewer)
	g.FightingQueue = make(map[string]bool)
	return &g
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
	http.HandleFunc("/", g.ServeAssets)
	go http.ListenAndServe(":"+Port, nil)
	log.I("http chat overlay started on port " + Port)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
