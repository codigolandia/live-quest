package oauth

import (
	"fmt"
	"github.com/codigolandia/live-quest/log"
	"net/http"
	"testing"
	"time"
)

func TestRandomState(t *testing.T) {
	str := randomState()
	t.Logf("Got random string: %#v", str)
	if str == "" {
		t.Error("wanted random string, got empty string")
	}

	if len(str) != 32 {
		t.Errorf("wanted 10 byte string, got %d", len(str))
	}
}

func TestWaitForCode(t *testing.T) {
	log.LogLevel = log.All
	state := randomState()
	var code string

	go func() {
		code = waitForAuthCode(state)
		log.D("test: received code: %v", code)
	}()
	time.Sleep(1 * time.Second)

	resp, err := http.Get(fmt.Sprintf("http://%s/oauth/youtube?state=%s&code=AUTHCODE",
		oauthCallbackAddr, state))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	time.Sleep(1 * time.Second)
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: %v %v", resp.StatusCode, resp.Status)
	}
	if code != "AUTHCODE" {
		t.Errorf("unexpected code detected: %v", code)
	}
}
