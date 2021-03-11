package server

import (
	"context"
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"log"
)

var Connections = make(map[string]*ConnectedClient)

type ConnectedClient struct {
	uid  string
	conn CleanableConnection
}

type CleanableConnection interface {
	GetConnection() io.ReadWriteCloser
	CleanUp(uid string) error
}

type WebsocketIO interface {
	WriteMessage(c *CleanableConnection, b []byte) error
	ReadMessage(c *CleanableConnection) ([]byte, error)
}

// make websocket io funcs mockable
var Server WebsocketIO

type websock struct{}

func init() {
	Server = websock{}
}

// wrap gobwas write
func (w websock) WriteMessage(c *CleanableConnection, b []byte) error {
	return wsutil.WriteServerMessage((*c).GetConnection(), ws.OpText, b)
}

// wrap gobwas read
func (w websock) ReadMessage(c *CleanableConnection) ([]byte, error) {
	read, _, err := wsutil.ReadClientData((*c).GetConnection())
	return read, err
}

// creates new connected client and registers in connections lookup map
func NewConnection(uid string, conn CleanableConnection) *ConnectedClient {
	c := &ConnectedClient{
		uid:  uid,
		conn: conn,
	}
	Connections[uid] = c
	return c
}

// write to websocket
func (c *ConnectedClient) Write(b []byte) error {
	if b == nil || len(b) == 0 {
		return errors.New("cannot write empty byte slice")
	}
	if c.conn == nil {
		return errors.New("connection is nil during write")
	}
	err := Server.WriteMessage(&c.conn, b)
	if err != nil {
		log.Println("error writing socket,", err)
		if err == io.EOF {
			if _, ok := Connections[c.uid]; ok {
				c.conn.CleanUp(c.uid)
				delete(Connections, c.uid)
			}
		}
		return err
	}
	return nil
}

// Msg describes what is read for who
type Msg struct {
	From string
	Data []byte
}

// read loop for websocket
func (c *ConnectedClient) Read(ctx context.Context, msgCh chan Msg) error {
	if c.conn == nil {
		return errors.New("connection is nil during write")
	}

	go func() {
		defer c.conn.GetConnection().Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			m, err := Server.ReadMessage(&c.conn)
			if err != nil {
				log.Println("error in ws read:", err.Error())
				if _, ok := Connections[c.uid]; ok {
					c.conn.CleanUp(c.uid)
					delete(Connections, c.uid)
				}
				log.Println("ending connection")
				return
			}

			if msgCh != nil {
				msgCh <- Msg{From: c.uid, Data: m}
			}
		}
	}()

	return nil
}

func (c *ConnectedClient) Close() {
	if c.conn != nil {
		_ = c.conn.CleanUp(c.uid)
		_ = c.conn.GetConnection().Close()
	}
}
