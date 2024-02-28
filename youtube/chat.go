package youtube

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
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

	unreadMu sync.Mutex
	unread   []message.Message
}

func New() (*Client, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("youtube: environment variable unset: YOUTUBE_API_KEY")
	}

	ctx := context.Background()
	svc, err := yt.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("youtube: error initializing Youtube service: %v", err)
	}

	if LiveId == "" {
		return nil, fmt.Errorf("youtube: missing flag LiveID")
	}

	c := &Client{
		hc:     http.Client{},
		svc:    svc,
		apiKey: apiKey,
		unread: make([]message.Message, 0, 10),
	}
	c.goReadTheMessages()

	return c, nil
}

func (c *Client) currentStream() string {
	// TODO: automatizar a busca pela live atual
	// Requer OAuth2 para acessar a live ativa.
	return LiveId
}

func (c *Client) loadChatId() string {
	if c.chatId != "" {
		return c.chatId
	}

	liveId := c.currentStream()
	log.I("loading chat with ID %v", liveId)
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
		log.E("error parsing chatID: %v", err)
	}
	return c.chatId
}

func (c *Client) FetchMessages() (msg []message.Message) {
	c.unreadMu.Lock()
	defer c.unreadMu.Unlock()

	if len(c.unread) < 0 {
		return
	}
	msg = make([]message.Message, len(c.unread))
	copy(msg, c.unread)
	c.unread = make([]message.Message, 0, 10)
	return msg
}

func (c *Client) goReadTheMessages() {
	go func() {
		for {
			// Avoid repeating many requests if an error happens
			c.lastFetchTime = time.Now()
			c.pollingInterval = 10 * time.Second

			fields := []string{"authorDetails,snippet"}
			req := c.svc.LiveChatMessages.List(c.loadChatId(), fields)
			req.PageToken(c.nextPageToken)
			resp, err := req.Do()
			if err != nil {
				log.E("error loading messages: %v", err)
				return
			}
			for i := range resp.Items {
				item := resp.Items[i]
				timeStamp, err := time.Parse(time.RFC3339Nano, item.Snippet.PublishedAt)
				if err != nil {
					log.E("unable to parse %v as timestamp: %v",
						item.Snippet.PublishedAt, err)
					timeStamp = time.Now()
				}
				c.unreadMu.Lock()
				c.unread = append(c.unread, message.Message{
					UID:       item.AuthorDetails.ChannelId,
					Author:    item.AuthorDetails.DisplayName,
					Text:      item.Snippet.DisplayMessage,
					Timestamp: timeStamp,
					Platform:  message.PlatformYoutube,
				})
				c.unreadMu.Unlock()
			}

			c.nextPageToken = resp.NextPageToken
			d := time.Duration(resp.PollingIntervalMillis) * time.Millisecond
			c.pollingInterval = d

			// Wait for pollingInterval to be passed before calling again.
			time.Sleep(max(c.pollingInterval, 5*time.Second))
			log.D("waiting for messages")
		}
	}()
}

func (c *Client) NextPageToken() string {
	if c == nil {
		return ""
	}
	return c.nextPageToken
}

func (c *Client) SetPageToken(t string) {
	if c == nil {
		return
	}
	c.nextPageToken = t
}
