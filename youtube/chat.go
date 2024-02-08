package youtube

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/codigolandia/jogo-da-live/log"
	"github.com/codigolandia/jogo-da-live/message"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

var LiveId = ""

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

	liveId := c.currentStream()
	log.I("carregando o chat ID da live %v", liveId)
	req := c.svc.Videos.List([]string{"liveStreamingDetails"}).Id(liveId)
	callback := func(resp *yt.VideoListResponse) error {
		for i := range resp.Items {
			v := resp.Items[i]
			c.chatId = v.LiveStreamingDetails.ActiveLiveChatId
			log.I("> chatId: %v", c.chatId)
			break
		}
		return nil
	}
	if err := req.Pages(context.Background(), callback); err != nil {
		log.E("erro a obter chat ID: %v", err)
	}
	return c.chatId
}

func (c *Client) FetchMessages() (msg []message.Message) {
	if c.nextPageToken != "" && time.Since(c.lastFetchTime) < c.pollingInterval {
		return msg
	}

	req := c.svc.LiveChatMessages.List(c.loadChatId(), []string{"authorDetails,snippet"})
	req.PageToken(c.nextPageToken)
	resp, err := req.Do()
	if err != nil {
		log.E("erro obtendo mensagens: %v", err)
		return msg
	}
	for i := range resp.Items {
		item := resp.Items[i]
		timeStamp, err := time.Parse(time.RFC3339Nano, item.Snippet.PublishedAt)
		if err != nil {
			log.E("erro ao interpretar %v como um timestamp: %v",
				item.Snippet.PublishedAt, err)
			timeStamp = time.Now()
		}
		msg = append(msg, message.Message{
			UID:       item.AuthorDetails.ChannelId,
			Author:    item.AuthorDetails.DisplayName,
			Text:      item.Snippet.DisplayMessage,
			Timestamp: timeStamp,
			Platform:  message.PlatformYoutube,
		})
	}

	c.nextPageToken = resp.NextPageToken
	d := time.Duration(resp.PollingIntervalMillis) * time.Millisecond
	c.pollingInterval = d
	c.lastFetchTime = time.Now()

	return msg
}
