package assets

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

var Bundles map[string]Bundle

type Bundle map[string]*Animation

func init() {
	Bundles = make(map[string]Bundle)
	Bundles["gopher"] = MustLoadAnimation("gopher")
}

type Animation struct {
	Frames []*ebiten.Image
	Skins  []*ebiten.Image
}

func LoadAnimation(path string) (bundle Bundle, err error) {
	bundle = make(map[string]*Animation)
	frames, err := fs.Glob(Assets, "animations/"+path+"/*/frame*.png")
	if err != nil {
		return nil, err
	}
	skins, err := fs.Glob(Assets, "animations/"+path+"/*/skin*.png")
	if err != nil {
		return nil, err
	}
	matches := []string{}
	matches = append(matches, frames...)
	matches = append(matches, skins...)
	for _, f := range matches {
		pathParts := strings.Split(f, "/")
		if len(pathParts) != 4 {
			return nil, fmt.Errorf("animation: invalid path parts: %#v", pathParts)
		}
		anim, fileName := pathParts[2], pathParts[3]
		a, ok := bundle[anim]
		if !ok {
			a = &Animation{}
		}
		if strings.HasPrefix(fileName, "frame") {
			img := LoadEbitenImg(f, true)
			a.Frames = append(a.Frames, img)
		} else {
			img := LoadEbitenImg(f, false)
			a.Skins = append(a.Skins, img)
		}
		bundle[anim] = a
	}
	return bundle, nil
}

func MustLoadAnimation(path string) Bundle {
	bundle, err := LoadAnimation(path)
	if err != nil {
		panic(err)
	}
	return bundle
}
