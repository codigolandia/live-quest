package twitch

import (
	"flag"
	"fmt"
	"testing"
)

var integrationTest bool

func init() {
	flag.BoolVar(&integrationTest, "it", false, "Run integration tests")
	Channel = "codigolandia"
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
