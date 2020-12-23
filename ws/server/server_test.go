package server

import (
	"context"
	"errors"
	"github.com/mousybusiness/go-web/ws/wstest"
	"io"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

type MockServer struct {
	WriteErr  error
	ReadBytes []byte
	ReadErr   error
}

func (m MockServer) WriteMessage(c *CleanableConnection, b []byte) error {
	return m.WriteErr
}

func (m MockServer) ReadMessage(c *CleanableConnection) ([]byte, error) {
	return m.ReadBytes, m.ReadErr
}

func TestNewConnection(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs

	conn := wstest.MockCleanConn{}
	uid := "stub"
	c := NewConnection(uid, conn)
	if c == nil {
		t.Fatalf("connection nil")
	}

	if v, ok := Connections[uid]; !ok {
		t.Fatalf("connection was not added to connections lookup")
	} else {
		if v.uid != uid {
			t.Fatalf("invalid uid assigned to connection")
		}
	}
}

func TestWrite(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs

	Server = MockServer{}

	conn := wstest.MockCleanConn{}
	uid := "stub"
	c := NewConnection(uid, conn)

	// nil data
	err := c.Write(nil)
	checkErrNil(t, err)

	// empty data
	err = c.Write([]byte{})
	checkErrNil(t, err)

	// happy path write
	err = c.Write([]byte{1, 2, 3})
	checkErr(t, err)

	// error during write - non EOF
	Server = MockServer{WriteErr: errors.New("error during write")}
	err = c.Write([]byte{1, 2, 3})
	checkErrNil(t, err)
	if _, ok := Connections[uid]; !ok {
		t.Fatalf("connection shouldnt be removed on non-EOF error")
	}

	// EOF during write - client disconnected
	Server = MockServer{WriteErr: io.EOF}
	err = c.Write([]byte{1, 2, 3})
	checkErrNil(t, err)
	if _, ok := Connections[uid]; ok {
		t.Fatalf("connection should be removed if EOF")
	}
}

func TestRead(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs

	conn := wstest.MockCleanConn{
		Conn: wstest.MockRWCloser{},
	}
	uid := "stub-uid"
	c := NewConnection(uid, conn)

	ctx, cancel := context.WithCancel(context.Background())

	// happy path
	Server = MockServer{ReadBytes: []byte("stub")}
	timeout := time.NewTimer(time.Millisecond * 100)

	msgCh := make(chan Msg)
	c.Read(ctx, msgCh)

	select {
	case <-timeout.C:
		t.Fatalf("should return before timeout")
	case msg := <-msgCh:
		s := string(msg.Data)
		if s != "stub" {
			t.Fatalf("invalid response from msg channel")
		}

		if msg.From != uid {
			t.Fatalf("invalid uid attributed to msg")
		}
	}

	cancel()
	timeout.Reset(time.Millisecond * 10)

	// ensure loop remains open if nil messages are returned
	Server = MockServer{}
	msgCh = make(chan Msg)
	ctx, cancel = context.WithCancel(context.Background())
	timeout.Reset(time.Millisecond * 10)
	c.Read(ctx, msgCh)

	select {
	case <-ctx.Done():
		t.Fatalf("context shouldnt be cancelled")
	case <-timeout.C:
	}

	// cancel context
	timeout.Reset(time.Millisecond * 100)
	msgCh = make(chan Msg)
	ctx, cancel = context.WithCancel(context.Background())
	c.Read(ctx, msgCh)
	cancel() // cancel context`

	select {
	case <-ctx.Done():
	case <-timeout.C:
		t.Fatalf("context should have cancelled before timeout")
	}

	// error during read
	log.SetOutput(ioutil.Discard) // throw out error logs
	Server = MockServer{ReadBytes: []byte("stub"), ReadErr: errors.New("error during read")}
	msgCh = make(chan Msg)
	ctx, cancel = context.WithCancel(context.Background())
	timeout.Reset(time.Millisecond * 50)
	c.Read(ctx, msgCh)
	select {
	case <-ctx.Done():
	case <-timeout.C:
	case <-msgCh:
		t.Fatalf("shouldnt send to channel if error")
	}

	if _, ok := Connections[uid]; ok {
		t.Fatalf("should remove connection if error")
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
