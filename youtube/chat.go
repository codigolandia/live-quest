package youtube

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

var LiveId = ""

func I(msg string, args ...any) {
	log.Printf("youtube: INFO: "+msg, args...)
}

func D(msg string, args ...any) {
	log.Printf("youtube: DEBG: "+msg, args...)
}

func E(msg string, args ...any) {
	log.Printf("youtube: ERRO: "+msg, args...)
}

type Client struct {
	hc     http.Client
	svc    *yt.Service
	apiKey string

	chatId          string
	nextPageToken   string
	pollingInterval time.Duration
	lastFetchTime   time.Time
}

func New() (*Client, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("youtube: defina a variável YOUTUBE_API_KEY")
	}

	ctx := context.Background()
	svc, err := yt.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("youtube: erro inicializando serviço do Youtube: %v", err)
	}

	return &Client{
		hc:     http.Client{},
		svc:    svc,
		apiKey: apiKey,
	}, nil
}

type ListChatMessagesResponse struct {
	NextPageToken   string `json:"nextPageToken"`
	PollingInterval int    `json:"pollingIntervalMillis"`

	Items []Message `json:"items"`
}

type Message struct {
	AuthorDetails struct {
		DisplayName string `json:"displayName"`
	} `json:"authorDetails"`
}

type Video struct {
	LiveStreamingDetails struct {
		ChatId string `json:"activeLiveChatId"`
	} `json:"liveStreamingDetails"`
}

type ListVideoResponse struct {
	Items []Video `json:"items"`
}

func (c *Client) currentStream() string {
	if LiveId == "" {
		panic("Missing live id")
	}
	// TODO: automatizar a busca pela live atual
	// Requer OAuth2 para acessar a live ativa.
	return LiveId
}

func (c *Client) loadChatId() string {
	if c.chatId != "" {
		return c.chatId
	}

	req := c.svc.Videos.List([]string{"liveStreamingDetails"}).Id(c.currentStream())
	callback := func(resp *yt.VideoListResponse) error {
		for i := range resp.Items {
			v := resp.Items[i]
			c.chatId = v.LiveStreamingDetails.ActiveLiveChatId
			I("> chatId: %v", c.chatId)
			break
		}
		return nil
	}
	if err := req.Pages(context.Background(), callback); err != nil {
		E("erro a obter chat ID: %v", err)
	}
	return c.chatId
}

func (c *Client) FetchMessages() (msg []*yt.LiveChatMessage) {
	if c.nextPageToken != "" && time.Since(c.lastFetchTime) < c.pollingInterval {
		return msg
	}

	req := c.svc.LiveChatMessages.List(c.loadChatId(), []string{"authorDetails"})
	req.PageToken(c.nextPageToken)
	resp, err := req.Do()
	if err != nil {
		E("erro obtendo mensagens: %v", err)
		return msg
	}
	for i := range resp.Items {
		msg = append(msg, resp.Items[i])
	}

	c.nextPageToken = resp.NextPageToken
	d := time.Duration(resp.PollingIntervalMillis) * time.Millisecond
	c.pollingInterval = d
	c.lastFetchTime = time.Now()

	return msg
}
