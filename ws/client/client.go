package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	errs "github.com/pkg/errors"
	"log"
	"net/http"
	"net/url"
)

type websocketIO interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
}

type connection struct {
	Name string
	Conn websocketIO
}

type dialer interface {
	Dial(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

// creates new connection
func NewConnection(d dialer, secure bool, name, host, path, token string) (*connection, error) {
	if name == "" || host == "" || path == "" || token == "" {
		return nil, errors.New(fmt.Sprintf("invalid host or path, host: %s, path: %s", host, path))
	}

	scheme := "ws"
	if secure {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	log.Printf("connecting to %s", u.String())
	h := http.Header{}
	h.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c, _, err := d.Dial(u.String(), h)
	if err != nil {
		return nil, errs.Wrap(err, "failed to dial websocket")
	}

	log.Println("connected!")
	conn := &connection{
		Name: name,
		Conn: c,
	}
	return conn, nil
}

//  write to websocket
func (c *connection) Write(b []byte) error {
	return c.Conn.WriteMessage(websocket.TextMessage, b)
}

// read loop for wesocket
func (c *connection) Read(ctx context.Context, msgCh chan []byte) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, m, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("error in client read, was connection was close by server?")
				close(msgCh)
				return
			}
			msgCh <- m
		}
	}()
}
