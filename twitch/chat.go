package twitch

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/codigolandia/live-quest/log"
	"github.com/codigolandia/live-quest/message"
)

var addr = "irc.chat.twitch.tv:6667"
var Channel = ""

func init() {
	flag.StringVar(&Channel, "twitch-channel", "codigolandia", "The Twitch channel to connect to.")
}

type Client struct {
	conn   net.Conn
	reader *textproto.Reader

	unreadMu sync.Mutex
	unread   []message.Message
}

func New() (c *Client, err error) {
	if Channel == "" {
		return nil, fmt.Errorf("twitch: no channel informed; missing --twitch-channel parameter?")
	}
	log.I("connecting to twitch IRC server at %s", addr)
	c = &Client{}
	c.unread = make([]message.Message, 0, 10)
	c.conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.E("error connecting: %v", err)
		return nil, err
	}
	c.reader = textproto.NewReader(bufio.NewReader(c.conn))
	c.goReadTheMessages()
	log.D("tcp connection stablished")

	log.D("sending NICK command (err=%v)", c.send("NICK justinfan12345"))
	log.D("joining channel (err=%v)", c.send("JOIN #"+Channel))
	return c, nil
}

func parseAuthor(src string) string {
	parts := strings.Split(src, "!")
	return strings.ReplaceAll(parts[0], ":", "")
}

func (c *Client) goReadTheMessages() {
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			log.D("waiting for new messages...")
			msg, err := c.reader.ReadLine()
			if err != nil {
				log.E("error reading new message: %v", err)
				continue
			}
			// :foo!foo@foo.tmi.twitch.tv PRIVMSG #bar :bleedPurple
			fields := strings.Fields(msg)
			if len(fields) == 2 && fields[0] == "PING" {
				// PING :tmi.twitch.tv
				log.I("sending PONG")
				c.send("PONG")
				continue
			}
			if len(fields) < 4 {
				log.E("ignoring: %v", msg)
				continue
			}
			uid, cmd, ch := fields[0], fields[1], fields[2]
			switch cmd {
			case "PRIVMSG":
				author := parseAuthor(uid)
				msgText := strings.Join(fields[3:], " ")
				if strings.HasPrefix(msgText, ":") {
					msgText = msgText[1:]
				}
				c.unreadMu.Lock()
				c.unread = append(c.unread, message.Message{
					UID:       uid,
					Author:    author,
					Text:      msgText,
					Timestamp: time.Now(),
					Platform:  message.PlatformTwitch,
				})
				c.unreadMu.Unlock()
				log.D("new message from '%v' at %v: %v", author, ch, msgText)
			default:
				log.D("ignoring: %v:", msg)
			}
		}
	}()
}

func (c *Client) recv() (msg string, err error) {
	return c.reader.ReadLine()
}

func (c *Client) send(msg string) error {
	n, err := fmt.Fprintf(c.conn, msg+"\n")
	log.D("%d bytes sent (err=%v)", n, err)
	return err
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

func (c *Client) Close() {
	log.D("closing connection (err=%v)", c.conn.Close())
}
