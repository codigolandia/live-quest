package assets_test

import (
	"io"
	"testing"

	"github.com/codigolandia/live-quest/assets"
)

func TestReadFiles(t *testing.T) {
	testCases := []struct {
		args  string
		bytes int
	}{
		{"img/hp_bar_bg.png", 616},
		{"img/hp_bar_fg.png", 584},
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
