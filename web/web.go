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

// authenticated GET helper using TOKEN env variable
func AGet(url string, timeout time.Duration, headers ...KV) (int, []byte, error) {
	return Get(url, timeout, append(headers, getAuthKV())...)
}

// http GET helper
func Get(url string, timeout time.Duration, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}

	return do(req, timeout, headers...)
}

// authenticated PATCH helper using TOKEN env variable
func APatch(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	return Patch(url, timeout, b, append(headers, getAuthKV())...)
}

// http PATCH helper
func Patch(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(b))
	if err != nil {
		return 0, nil, err
	}

	return do(req, timeout, headers...)
}

// authenticated POST helper using TOKEN env variable
func APost(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	return Post(url, timeout, b, append(headers, getAuthKV())...)
}

// http POST helper
func Post(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return 0, nil, err
	}

	return do(req, timeout, headers...)
}

// authenticated PUT helper using TOKEN env variable
func APut(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	return Post(url, timeout, b, append(headers, getAuthKV())...)
}

// http PUT helper
func Put(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return 0, nil, err
	}

	return do(req, timeout, headers...)
}

// authenticated DELETE helper using TOKEN env variable
func ADelete(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	return Delete(url, timeout, b, append(headers, getAuthKV())...)
}

// http DELETE helper
func Delete(url string, timeout time.Duration, b []byte, headers ...KV) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(b))
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

	// a value of 0 means no timeout
	if timeout.Minutes() != 0 {
		Client.SetTimeout(timeout)
	}
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
