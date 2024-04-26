package main

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/codigolandia/live-quest/assets"
)

func (g *Game) ServeChat(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	if err := enc.Encode(g.ChatHistory); err != nil {
		http.Error(w, "game: error serializing chat: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

var dynamicHandler = http.FileServer(http.Dir("./assets/web"))
var embeddedHandler http.Handler

func init() {
	web, err := fs.Sub(assets.Assets, "web")
	if err != nil {
		panic(err)
	}
	embeddedHandler = http.FileServer(http.FS(web))
}

func (g *Game) ServeAssets(w http.ResponseWriter, r *http.Request) {
	if HttpHotReload {
		dynamicHandler.ServeHTTP(w, r)
	} else {
		embeddedHandler.ServeHTTP(w, r)
	}
	w.Header().Set("cache-control", "no-cache")
}
