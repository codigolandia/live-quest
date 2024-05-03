package oauth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/codigolandia/live-quest/log"
	"github.com/codigolandia/live-quest/message"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var oauthCallbackAddr = "localhost:8089"

type persistentTokenSource struct {
	provider string
	wrapped  oauth2.TokenSource
	token    *oauth2.Token
}

func (p *persistentTokenSource) Token() (t *oauth2.Token, err error) {
	newToken, err := p.wrapped.Token()
	if err != nil {
		return nil, err
	}
	p.token = newToken

	tokenFile := tokenFileName(p.provider)
	fd, err := os.OpenFile(tokenFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.E("token: unable to save token at %s: %v", tokenFile, err)
		return p.token, nil
	}
	defer fd.Close()
	if err := json.NewEncoder(fd).Encode(newToken); err != nil {
		log.E("token: unable to serialize token: %v", err)
	}
	return p.token, err
}

func NewTokenSource(provider string) (ts oauth2.TokenSource, err error) {
	if ts, ok := isAuthenticated(provider); ok {
		return ts, nil
	}

	ctx := context.Background()
	config := configForProvider(provider)

	// 1. Redirect to url
	state := randomState()
	url := config.AuthCodeURL(state)
	fmt.Printf("Authentication required for %v", provider)
	fmt.Printf("Open in your browser:\n\n%v\n\n", url)
	// 2. Exchange code for token
	code := waitForAuthCode(state)
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	// 3. Return custom token source
	origTokenSource := config.TokenSource(ctx, token)
	ts = &persistentTokenSource{
		wrapped:  origTokenSource,
		token:    token,
		provider: provider,
	}
	return ts, nil
}

func isAuthenticated(provider string) (ts oauth2.TokenSource, ok bool) {
	tokenFile := tokenFileName(provider)
	fd, err := os.Open(tokenFile)
	if err != nil {
		log.W("token: error loading previous token at %v: %v", tokenFile, err)
		return nil, false
	}
	defer fd.Close()
	token := &oauth2.Token{}
	if err := json.NewDecoder(fd).Decode(token); err != nil {
		log.W("token: error deserializing token: %v", err)
		return nil, false
	}

	config := configForProvider(provider)
	ts = config.TokenSource(context.Background(), token)
	return ts, true
}

func configForProvider(provider string) oauth2.Config {
	switch provider {
	case message.PlatformYoutube:
		return oauth2.Config{
			ClientID:     os.Getenv("YOUTUBE_CLIENT_ID"),
			ClientSecret: os.Getenv("YOUTUBE_CLIENT_SECRET"),
			Endpoint:     endpoints.Google,
			RedirectURL:  fmt.Sprintf("http://%s/oauth/youtube", oauthCallbackAddr),
			Scopes: []string{
				"https://www.googleapis.com/auth/youtube",
			},
		}
	case message.PlatformTwitch:
		return oauth2.Config{
			ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
			ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
			Endpoint:     endpoints.Twitch,
			RedirectURL:  fmt.Sprintf("http://%s/oauth/twitch", oauthCallbackAddr),
			Scopes: []string{
				"chat:read",
				"chat:edit",
			},
		}
	}
	panic(fmt.Errorf("unexepected provider: %v", provider))
}

func randomState() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Should never happen
		panic(err)
	}
	return fmt.Sprintf("%0x", b)
}

func waitForAuthCode(state string) (code string) {
	sem := make(chan struct{})
	s := http.Server{
		Addr: oauthCallbackAddr,
	}
	// TODO(ronoaldo): configure timeout
	shutdown := func() {
		log.D("oauth: shutting down server: %v", s.Shutdown(context.Background()))
		sem <- struct{}{}
	}
	s.IdleTimeout = time.Second * 15
	s.ReadTimeout = time.Second * 15
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qs := r.URL.Query()
		receivedState := qs.Get("state")
		if state != receivedState {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}
		if qs.Get("error") != "" {
			log.E("oauth: error received when exchanging code: %v (%v)",
				qs.Get("error"), qs.Get("error_description"))
		}

		code = qs.Get("code")
		if code == "" {
			log.E("oauth: missing code parameter")
		}
		log.D("oauth: received code %v", mask(code))
		w.Write([]byte("OK"))
		go shutdown()
	})
	log.D("oauth: initializing auth code server ...")
	log.D("oauth: server finished with status: %v", s.ListenAndServe())
	<-sem
	return code
}

func mask(src string) string {
	if len(src) < 1 {
		return "*"
	}
	if len(src) < 5 {
		return string(src[0]) + "****"
	}
	return src[:3] + "****"
}

func tokenFileName(provider string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.W("token: unable to detect user home, using current directory")
		home = "."
	}
	return filepath.Join(home, fmt.Sprintf(".live-quest.%s.json", provider))
}
