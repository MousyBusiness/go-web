package main

import (
	"github.com/mousybusiness/go-web/web"
	"log"
	"os"
	"os/signal"
	"time"
)

func MockedFunction() (int, error) {
	code, _, err := web.Get("https://www.google.com", time.Second*3)
	return code, err
}

func main() {
	// call function that can be mocked
	code, err := MockedFunction()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("response code was", code)

	// Wait for Control C to exit
	block := make(chan os.Signal, 1)
	signal.Notify(block, os.Interrupt)

	// Block until signal is received
	<-block
}
