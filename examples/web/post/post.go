package main

import (
	"fmt"
	"github.com/mousybusiness/go-web/web"
	"log"
	"os"
	"time"
)

func main() {
	if os.Getenv("TOKEN") == "" {
		log.Fatalln("require token environment variable")
	}

	url := fmt.Sprintf("http://%s/test", "example.com")

	// by using APost (Authenticated Post)  KV{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TOKEN"))} will be added as a header
	code, bytes, err := web.APost(url, time.Second*10, []byte("{}"))

	log.Println(code, string(bytes), err)
}
