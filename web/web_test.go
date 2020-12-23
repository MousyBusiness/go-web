package web

import (
	"bytes"
	"errors"
	"github.com/mousybusiness/go-web/web/webtest"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	log.SetOutput(ioutil.Discard) // discard logs
	Client = webtest.MockClient{}
	req := &http.Request{Header: make(map[string][]string)}
	webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("stub"))), StatusCode: 200}, nil
	}

	// happy path - no
	code, body, err := do(req, time.Millisecond*100)
	checkErr(t, err)

	if code != 200 {
		t.Fatalf("do http status code; wanted: %v, got: %v", 200, code)
	}

	if string(body) != "stub" {
		t.Fatalf("do http body response; wanted: %v, got: %v", "stub", string(body))
	}

	webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("stub")))}, errors.New("error in http")
	}

	// error in http call
	_, _, err = do(req, time.Millisecond*100)
	checkErrNil(t, err)

	// check headers
	var h map[string][]string
	webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
		h = req.Header
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("stub")))}, nil
	}

	do(req, time.Millisecond*100, KV{"Content-Type", "stub"}, KV{"X-Api-Key", "123"})

	if v, ok := h["Content-Type"]; !ok {
		t.Fatalf("expecting content type header in request")
	} else {
		if v[0] != "stub" {
			t.Fatalf("header content type; want: %v, got: %v", "stub", v)
		}
	}

	if v, ok := h["X-Api-Key"]; !ok {
		t.Fatalf("expecting api key header in request")
	} else {
		if v[0] != "123" {
			t.Fatalf("header api key; want: %v, got: %v", "stub", v)
		}
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
