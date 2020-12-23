package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	SetTimeout(timeout time.Duration)
}

var Client HTTPClient

type client struct {
	c *http.Client
}

func init() {
	Client = client{
		c: &http.Client{},
	}
}

func (c client) Do(req *http.Request) (*http.Response, error) {
	return c.c.Do(req)
}

func (c client) SetTimeout(timeout time.Duration) {
	c.c.Timeout = timeout
}

type KV struct {
	Key   string
	Value string
}

// authenticated get helper using TOKEN env variable
func AGet(url string, timeout time.Duration, headers ...KV) (int, []byte, error) {
	return Get(url, timeout, append(headers, getAuthKV())...)
}

// http get helper
func Get(url string, timeout time.Duration, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, err
	}

	return do(req, timeout, headers...)
}

// authenticated post helper using TOKEN env variable
func APost(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	return Post(url, timeout, b, append(headers, getAuthKV())...)
}

// http post helper
func Post(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return 0, nil, err
	}

	return do(req, timeout, headers...)
}

func do(req *http.Request, timeout time.Duration, headers ...KV) (int, []byte, error) {
	req.Header.Set("Content-Type", "application/json") // default to json
	for _, v := range headers {
		req.Header.Set(v.Key, v.Value)
	}

	Client.SetTimeout(timeout)
	resp, err := Client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}

func getAuthKV() KV {
	return KV{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TOKEN"))}
}
