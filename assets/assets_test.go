package assets_test

import (
	"io"
	"testing"

	"github.com/codigolandia/jogo-da-live/assets"
)

func TestReadFiles(t *testing.T) {
	testCases := []struct {
		args  string
		bytes int
	}{
		{"img/gopher_standing.gif", 5323},
		{"img/gopher_standing.png", 2071},
		{"img/gopher_walking_left.gif", 7605},
		{"img/gopher_walking_left.png", 3673},
		{"img/gopher_walking_right.gif", 7559},
		{"img/gopher_walking_right.png", 3998},
	}

	for _, tc := range testCases {
		fd, err := assets.Assets.Open(tc.args)
		t.Logf("Read %v (err=%v)", tc.args, err)
		if err != nil {
			t.Fatalf("unexpected error openning image: %v", err)
		}

		b, err := io.ReadAll(fd)
		if err != nil {
			t.Fatalf("unexpected error while loading image bytes: %v", err)
		}

		if len(b) != tc.bytes {
			t.Errorf("unexpected byte size for file: got %v, want %v",
				len(b), tc.bytes)
		}
	}
}
