package main

import (
	"github.com/mousybusiness/go-web/web"
	"log"
	"time"
)

func main() {
	url := "http://metadata.google.internal/computeMetadata/v1/project/project-id"

	code, bytes, err := web.Get(url, time.Second*2, web.KV{"Metadata-Flavor", "Google"})

	log.Println(code, string(bytes), err)
}
