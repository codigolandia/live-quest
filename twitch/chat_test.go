package twitch

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/codigolandia/live-quest/log"
)

var integrationTest bool

func init() {
	flag.BoolVar(&integrationTest, "it", false, "Run integration tests")
}

func TestNewClient(t *testing.T) {
	if !integrationTest {
		t.Skip("Integration tests disabled, skipping...")
	}
	log.LogLevel = log.Debug
	c, err := New()
	if err != nil {
		t.Errorf("error initializing connection %v", err)
	}
	time.Sleep(60 * time.Second)
	defer c.Close()
}

func TestParseAuthor(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{":codigolandia!codigolandia@codigolandia.tmi.twitch.tv:", "codigolandia"},
		{":rodinei!rodinei@codigolandia.tmi.twitch.tv:", "rodinei"},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case#%d", i), func(t *testing.T) {
			res := parseAuthor(tc.input)
			t.Logf("%s => %s", tc.input, res)
			if res != tc.output {
				t.Errorf("nome do autor inválido: expected: %v, got: %v", tc.output, res)
			}
		})
	}
}
