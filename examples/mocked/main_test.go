package main

import (
	"errors"
	"github.com/mousybusiness/go-web/web"
	"github.com/mousybusiness/go-web/web/webtest"
	"net/http"
	"testing"
)

func init() {
	web.Client = webtest.MockClient{}
}

func TestMockedFunction(t *testing.T) {
	// mock status code
	webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
		return webtest.MockResponse(400, "bad request", nil)
	}

	code, _ := MockedFunction()
	if code != 400 {
		t.Fatalf("incorrect status code returned; wanted: %v, got: %v", 400, code)
	}

	// mock error
	webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
		return webtest.MockResponse(0, "errorororor", errors.New("an error"))
	}

	_, err := MockedFunction()
	if err == nil {
		t.Fatalf("expected error")
	}

}
