package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	errs "github.com/pkg/errors"
	"net/http"
	"net/url"
)

type websocketIO interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
}

type Dialer interface {
	Dial(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

type Connection struct {
	Name string
	Conn websocketIO
}

//  write to websocket
func (c *Connection) Write(b []byte) error {
	return c.Conn.WriteMessage(websocket.TextMessage, b)
}

// read loop for wesocket
func (c *Connection) Read(ctx context.Context, msgCh chan []byte) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, m, err := c.Conn.ReadMessage()
			if err != nil {
				close(msgCh)
				return
			}
			msgCh <- m
		}
	}()
}

// creates new Connection
func NewConnection(d Dialer, secure bool, name, host, path, token string) (*Connection, error) {
	if name == "" || host == "" || path == "" || token == "" {
		return nil, errors.New(fmt.Sprintf("invalid host or path, host: %s, path: %s", host, path))
	}

	scheme := "ws"
	if secure {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	h := http.Header{}
	h.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c, _, err := d.Dial(u.String(), h)
	if err != nil {
		return nil, errs.Wrap(err, "failed to dial websocket")
	}

	conn := &Connection{
		Name: name,
		Conn: c,
	}
	return conn, nil
}
