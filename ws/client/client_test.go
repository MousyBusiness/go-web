package client

import (
	"context"
	"errors"
	"github.com/mousybusiness/go-web/ws/wstest"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestNewConnection(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs

	tt := []struct {
		name  string
		host  string
		path  string
		token string
		isErr bool
	}{
		{"push-service", "0.0.0.0", "/echo", "123", false},
		{"", "0.0.0.0", "/echo", "123", true},
		{"push-service", "", "/echo", "123", true},
		{"push-service", "0.0.0.0", "", "123", true},
		{"push-service", "0.0.0.0", "/echo", "", true},
	}

	md := wstest.MockDialer{}

	for _, v := range tt {
		_, err := NewConnection(md, false, v.name, v.host, v.path, v.token)
		if !v.isErr {
			checkErr(t, err)
		} else {
			checkErrNil(t, err)
		}
	}

	// happy path
	md = wstest.MockDialer{}
	_, err := NewConnection(md, false, "stub", "stub", "stub", "stub")
	checkErr(t, err)

	// dial failed
	md = wstest.MockDialer{Err: errors.New("stub")}
	_, err = NewConnection(md, false, "stub", "stub", "stub", "stub")
	checkErrNil(t, err)
}

type WSConn struct {
	MsgType int
	Data    []byte
	Err     error
}

func (w WSConn) WriteMessage(messageType int, data []byte) error {
	if data == nil {
		return errors.New("stub")
	}
	return nil
}
func (w WSConn) ReadMessage() (messageType int, p []byte, err error) { return w.MsgType, w.Data, w.Err }

func TestWrite(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs

	w := WSConn{
		MsgType: 1,
		Data:    nil,
		Err:     nil,
	}
	conn := connection{
		Name: "stub",
		Conn: w,
	}

	// happy path
	err := conn.Write([]byte{1, 2, 3})
	checkErr(t, err)

	// nil data
	err = conn.Write(nil)
	checkErrNil(t, err)
}

func TestRead(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs

	// happy path
	w := WSConn{
		MsgType: 1,
		Data:    []byte{1, 2, 3},
		Err:     nil,
	}
	conn := connection{
		Name: "stub",
		Conn: w,
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan []byte)
	conn.Read(ctx, c)
	timer := time.NewTimer(time.Millisecond * 10)
	select {
	case <-c:
	case <-timer.C:
		t.Fatalf("expect channel result before timeout")
	}

	cancel()

	//error on read
	w = WSConn{
		MsgType: 1,
		Data:    []byte{1, 2, 3},
		Err:     errors.New("stub"),
	}
	conn = connection{
		Name: "stub",
		Conn: w,
	}

	ctx, cancel = context.WithCancel(context.Background())

	c = make(chan []byte)
	conn.Read(ctx, c)
	cancel()

	timer = time.NewTimer(time.Millisecond * 100)
	select {
	case <-ctx.Done():
		t.Log("context finished")
	case <-c:
		t.Fatalf("expect context finish")
	}

}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

func checkErrNil(t *testing.T, err error) {
	if err == nil {
		t.Helper()
		t.Fatal(err)
	}
}
