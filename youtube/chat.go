package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var LiveId = ""

func I(msg string, args ...any) {
	log.Printf("youtube: INFO: "+msg, args...)
}

func E(msg string, args ...any) {
	log.Printf("youtube: ERRO: "+msg, args...)
}

type Client struct {
	hc     http.Client
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

	return &Client{
		hc:     http.Client{},
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

func (c *Client) loadChatId() string {
	if c.chatId != "" {
		return c.chatId
	}

	url := "https://www.googleapis.com/youtube/v3/videos?"
	url += "&id=" + LiveId
	url += "&part=liveStreamingDetails"
	url += "&key=" + c.apiKey

	resp, err := c.hc.Get(url)
	if err != nil {
		E("erro ao obter chatid: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		E("erro ao obter chatid: %v %v", resp.StatusCode, string(b))
		return ""
	}

	videos := ListVideoResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&videos); err != nil {
		E("erro ao ler JSON: %v", err)
	}
	for _, v := range videos.Items {
		c.chatId = v.LiveStreamingDetails.ChatId
		I("found chatid %v", c.chatId)
		return c.chatId
	}
	return ""
}

func (c *Client) FetchMessages() (msg []Message) {
	if c.nextPageToken != "" && time.Since(c.lastFetchTime) < c.pollingInterval {
		return msg
	}
	url := "https://www.googleapis.com/youtube/v3/liveChat/messages?"
	url += "&part=authorDetails"
	url += "&pageToken=" + c.nextPageToken
	url += "&liveChatId=" + c.loadChatId()
	url += "&key=" + c.apiKey
	resp, err := c.hc.Get(url)
	if err != nil {
		E("erro obtendo mensagens: %v", err)
		return msg
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		E("erro ao obter mensagens: %v", string(b))
		return msg
	}

	chatMessages := ListChatMessagesResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&chatMessages); err != nil {
		E("erro decodificando a mensagem: %v", err)
		return msg
	}

	for i := range chatMessages.Items {
		msg = append(msg, chatMessages.Items[i])
	}

	c.nextPageToken = chatMessages.NextPageToken
	d := time.Duration(chatMessages.PollingInterval) * time.Millisecond
	c.pollingInterval = d
	c.lastFetchTime = time.Now()

	return msg
}
