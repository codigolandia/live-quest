package youtube

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/codigolandia/live-quest/log"
	"github.com/codigolandia/live-quest/message"
	"github.com/codigolandia/live-quest/youtube/proto"
	ytp "github.com/codigolandia/live-quest/youtube/proto"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

var (
	Channel         = ""
	LiveID          = ""
	IncludeUpcoming = false
)

func init() {
	flag.StringVar(&Channel, "youtube-channel", "", "The Youtube channel to connect to.")
	flag.StringVar(&LiveID, "youtube-stream", "", "The Youtube video ID of the livestream to connect to.")
	flag.BoolVar(&IncludeUpcoming, "youtube-include-scheduled", false, "If we should also include upcoming videos")
}

type Client struct {
	hc   http.Client
	svc  *yt.Service
	dsvc *ytp.DataClient
	ctx  context.Context

	chatId          string
	nextPageToken   string
	pollingInterval time.Duration
	lastFetchTime   time.Time

	unreadMu sync.Mutex
	unread   []message.Message
}

func New(nextPageToken string, tokenSource oauth2.TokenSource) (c *Client, err error) {
	if Channel == "" && LiveID == "" {
		return nil, fmt.Errorf("youtube: no channel informed; missing --youtube-channel parameter?")
	}
	log.I("youtube: continuing from page token: %v", nextPageToken)

	var (
		svc  *yt.Service
		dsvc *ytp.DataClient
	)

	ctx := context.Background()

	// TODO: Review if we will keep supporting API Key
	if tokenSource == nil {
		apiKey := os.Getenv("YOUTUBE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("youtube: environment variable unset: YOUTUBE_API_KEY")
		}
		svc, err = yt.NewService(ctx, option.WithAPIKey(apiKey))
	} else {
		svc, err = yt.NewService(ctx, option.WithTokenSource(tokenSource))
	}

	dsvc, err = ytp.NewClient(ctx, tokenSource)
	if err != nil {
		return nil, fmt.Errorf("youtube: error initializing gRPC client: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("youtube: error initializing Youtube service: %v", err)
	}

	c = &Client{
		hc:            http.Client{},
		svc:           svc,
		dsvc:          dsvc,
		ctx:           ctx,
		unread:        make([]message.Message, 0, 10),
		nextPageToken: nextPageToken,
	}
	c.goReadTheMessages()

	return c, nil
}

func (c *Client) currentStream() string {
	if LiveID != "" {
		return LiveID
	}
	channelResponse, err := c.svc.Channels.
		List([]string{"snippet"}).
		ForHandle(Channel).
		Do()
	if err != nil {
		log.E("youtube: error looking up the channel ID: %v", err)
		return ""
	}

	if len(channelResponse.Items) == 0 {
		log.E("youtube: channel not found for handle %v", Channel)
		return ""
	}
	channelID := channelResponse.Items[0].Id

	eventTypes := []string{"live"}
	if IncludeUpcoming {
		eventTypes = append(eventTypes, "upcoming")
	}
	for _, eventType := range eventTypes {
		searchResponse, err := c.svc.Search.
			List([]string{"snippet"}).
			ChannelId(channelID).
			EventType(eventType).
			Type("video").
			MaxResults(1).
			Do()
		if err != nil {
			log.E("youtube: error loading current live stream: %v", err)
			return ""
		}
		if len(searchResponse.Items) == 0 {
			log.E("youtube: no live streams found for channel %v (%v) [%v]", Channel, channelID, eventType)
		} else {
			LiveID = searchResponse.Items[0].Id.VideoId
			log.I("youtube: found %v stream with videoID: %v", eventType, LiveID)
			break
		}
	}

	return LiveID
}

func (c *Client) loadChatId() *string {
	if c.chatId != "" {
		return &c.chatId
	}

	liveId := c.currentStream()
	if liveId == "" {
		return &c.chatId
	}

	log.I("youtube: loading chat from videoID: %v", liveId)
	req := c.svc.Videos.List([]string{"liveStreamingDetails"}).Id(liveId)
	callback := func(resp *yt.VideoListResponse) error {
		for i := range resp.Items {
			v := resp.Items[i]
			c.chatId = v.LiveStreamingDetails.ActiveLiveChatId
			log.I("youtube: found chatId: %v", c.chatId)
			break
		}
		return nil
	}
	if err := req.Pages(context.Background(), callback); err != nil {
		log.E("error parsing chatID: %v", err)
	}
	return &c.chatId
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
	c.loadChatId()
	go func() {
		for {
			if c.loadChatId() == nil || *c.loadChatId() == "" {
				log.E("youtube: no live stream found. Waiting for 10s")
				time.Sleep(10 * time.Second)
				continue
			}

			// Avoid repeating many requests if an error happens
			c.lastFetchTime = time.Now()
			c.pollingInterval = 10 * time.Second

			req := proto.LiveChatMessageListRequest{
				Part:       []string{"snippet,authorDetails"},
				PageToken:  &c.nextPageToken,
				LiveChatId: c.loadChatId(),
			}
			streamList, err := c.dsvc.YT.StreamList(c.ctx, &req)

			if err == nil {
				for {
					resp, err := streamList.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.E("youtube: error receiving message: %v", err)
						time.Sleep(10 * time.Second)
						continue
					}
					c.nextPageToken = resp.GetNextPageToken()

					for _, item := range resp.Items {
						timeStamp, err := time.Parse(time.RFC3339Nano, *item.Snippet.PublishedAt)
						if err != nil {
							log.E("unable to parse %v as timestamp: %v",
								item.Snippet.PublishedAt, err)
							timeStamp = time.Now()
						}
						c.unreadMu.Lock()
						c.unread = append(c.unread, message.Message{
							UID:       *item.AuthorDetails.ChannelId,
							Author:    *item.AuthorDetails.DisplayName,
							Text:      *item.Snippet.DisplayMessage,
							Timestamp: timeStamp,
							Platform:  message.PlatformYoutube,
						})
						c.unreadMu.Unlock()
					}
				}
				// Wait for pollingInterval to be passed before calling again.
				d := time.Duration(1000) * time.Millisecond
				c.pollingInterval = d
				time.Sleep(max(c.pollingInterval, 3*time.Second))
			} else {
				log.E("youtube: rror loading messages: err=%v", err)
				time.Sleep(10 * time.Second)
			}
		}
	}()
}

func (c *Client) NextPageToken() string {
	if c == nil {
		return ""
	}
	return c.nextPageToken
}

func (c *Client) SendMessage(msg string) error {
	l := &yt.LiveChatMessage{
		Snippet: &yt.LiveChatMessageSnippet{
			LiveChatId: c.chatId,
			Type:       "textMessageEvent",
			TextMessageDetails: &yt.LiveChatTextMessageDetails{
				MessageText: msg,
			},
		},
	}
	_, err := c.svc.LiveChatMessages.Insert([]string{"snippet"}, l).Do()
	return err
}

func (c *Client) SetPageToken(t string) {
	if c == nil {
		return
	}
	c.nextPageToken = t
}
