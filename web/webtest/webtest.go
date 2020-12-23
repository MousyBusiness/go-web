package webtest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

var DoFunc func(req *http.Request) (*http.Response, error)

type MockClient struct {
}

func (m MockClient) Do(req *http.Request) (*http.Response, error) {
	return DoFunc(req)
}

func (m MockClient) SetTimeout(timeout time.Duration) {
}

func MockResponse(code int, body string, err error) (*http.Response, error) {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
	}, err
}
