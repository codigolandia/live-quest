package twitch

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"

	"github.com/codigolandia/jogo-da-live/log"
)

var addr = "irc.chat.twitch.tv:6667"

type Client struct {
	conn   net.Conn
	reader *textproto.Reader
}

func New() (c *Client, err error) {
	log.I("connectando ao Twitch em %s", addr)
	c = &Client{}
	c.conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.D("erro ao conectar: %v", err)
		return nil, err
	}
	c.reader = textproto.NewReader(bufio.NewReader(c.conn))
	c.goReadTheMessages()
	log.D("conexão tcp estabelecida")

	log.D("enviando nick (err=%v)", c.send("NICK justinfan12345"))
	log.D("entrando no servidor (err=%v)", c.send("JOIN #codigolandia"))
	return c, nil
}

func (c *Client) goReadTheMessages() {
	go func() {
		for {
			log.D("aguardando mensagem ...")
			msg, err := c.reader.ReadLine()
			if err != nil {
				log.E("erro ao ler mensagem: %v", err)
			}
			log.D("onMessage: received: %v: %v", msg, err)
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

func (c *Client) Close() {
	log.D("encerrando conexão (err=%v)", c.conn.Close())
}
