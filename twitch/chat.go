package twitch

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/codigolandia/jogo-da-live/log"
	"github.com/codigolandia/jogo-da-live/message"
)

var addr = "irc.chat.twitch.tv:6667"

type Client struct {
	conn   net.Conn
	reader *textproto.Reader

	unreadMu sync.Mutex
	unread   []message.Message
}

func New() (c *Client, err error) {
	log.I("connectando ao Twitch em %s", addr)
	c = &Client{}
	c.unread = make([]message.Message, 0, 10)
	c.conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.E("erro ao conectar: %v", err)
		return nil, err
	}
	c.reader = textproto.NewReader(bufio.NewReader(c.conn))
	c.goReadTheMessages()
	log.D("conexão tcp estabelecida")

	log.D("enviando nick (err=%v)", c.send("NICK justinfan12345"))
	log.D("entrando no servidor (err=%v)", c.send("JOIN #codigolandia"))
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
			log.D("aguardando mensagem ...")
			msg, err := c.reader.ReadLine()
			if err != nil {
				log.E("erro ao ler mensagem: %v", err)
				continue
			}
			fields := strings.Fields(msg)
			if len(fields) < 4 {
				log.E("mensagem inválida: %v", msg)
				continue
			}
			uid, cmd, ch := fields[0], fields[1], fields[2]
			switch cmd {
			case "PRIVMSG":
				author := parseAuthor(uid)
				msgText := strings.Join(fields[3:], " ")
				c.unreadMu.Lock()
				c.unread = append(c.unread, message.Message{
					UID:      uid,
					Author:   author,
					Text:     msgText,
					Platform: message.PlatformTwitch,
				})
				c.unreadMu.Unlock()
				log.D("nova mensagem de '%v' em %v: %v", author, ch, msgText)
			default:
				log.D("ignorando: %v:", msg)
			}
		}
	}()
}

func (c *Client) recv() (msg string, err error) {
	return c.reader.ReadLine()
}

func (c *Client) send(msg string) error {
	n, err := fmt.Fprintf(c.conn, msg+"\n")
	log.D("%d bytes enviados (err=%v)", n, err)
	return err
}

func (c *Client) FetchMessages() (msg []message.Message) {
	c.unreadMu.Lock()
	defer c.unreadMu.Unlock()

	if len(c.unread) < 0 {
		return
	}
	msg = make([]message.Message, len(c.unread))
	c.unread = make([]message.Message, 0, 10)
	return msg
}

func (c *Client) Close() {
	log.D("encerrando conexão (err=%v)", c.conn.Close())
}
