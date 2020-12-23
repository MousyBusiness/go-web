package wstest

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
)

type MockHappyConnection struct {
	Written []byte
	Msg     []byte
}

func (m *MockHappyConnection) Write(b []byte) error {
	m.Written = b
	return nil
}

func (m *MockHappyConnection) Read(ctx context.Context, msgCh chan []byte) {
	msgCh <- []byte("stub") // immediately receive a message
}

type MockUnhappyConnection struct {
	Written []byte
	Msg     []byte
}

func (m *MockUnhappyConnection) Write(b []byte) error {
	m.Written = nil
	return errors.New("write error")
}

func (m *MockUnhappyConnection) Read(ctx context.Context, msgCh chan []byte) {
	msgCh <- nil // immediately receive a message
}

type MockDialer struct {
	Conn *websocket.Conn
	Resp *http.Response
	Err  error
}

func (m MockDialer) Dial(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	return m.Conn, m.Resp, m.Err
}

type MockCleanConn struct {
	Conn       io.ReadWriteCloser
	CleanedUp  bool
	CleanupErr error
}

func (m MockCleanConn) GetConnection() io.ReadWriteCloser {
	return m.Conn
}

func (m MockCleanConn) CleanUp(uid string) error {
	m.CleanedUp = true
	return m.CleanupErr
}

type MockRWCloser struct{}

func (rw MockRWCloser) Read(p []byte) (n int, err error)  { return 0, nil }
func (rw MockRWCloser) Write(p []byte) (n int, err error) { return 0, nil }
func (rw MockRWCloser) Close() error                      { return nil }
