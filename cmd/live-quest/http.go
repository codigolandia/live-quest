package main

import (
	"io/fs"
	"net/http"

	"github.com/codigolandia/live-quest/assets"
)

func (g *Game) ServeChat(w http.ResponseWriter, r *http.Request) {
	tempFileMu.Lock()
	defer tempFileMu.Unlock()

	http.ServeFile(w, r, g.tempFile())
	w.Header().Set("cache-control", "no-cache")
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
