package twitch

import (
	"testing"
	"time"

	"github.com/codigolandia/jogo-da-live/log"
)

func TestNewClinet(t *testing.T) {
	log.LogLevel = log.Debug
	c, err := New()
	if err != nil {
		t.Errorf("error initializing connection %v", err)
	}
	time.Sleep(60 * time.Second)
	defer c.Close()
}
