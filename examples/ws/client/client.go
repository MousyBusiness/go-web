package main

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/mousybusiness/go-web/ws/client"
	"log"
	"os"
	"os/signal"
)

func handleMsg(msg []byte) error {
	log.Println(string(msg))
	return nil
}

func main() {
	ctx := context.Background()

	conn, _ := client.NewConnection(websocket.DefaultDialer, false, "my-connection-name", "http://myapp.com", "/signal", os.Getenv("TOKEN"), "")

	ch := make(chan []byte)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b, open := <-ch:
				if !open {
					log.Println("connection was closed, exiting read loop")
					return
				}
				err := handleMsg(b)
				if err != nil {
					log.Println("error during websocket read, ", err)
				}
			}
		}
	}()

	conn.Read(ctx, ch)

	// Wait for Control C to exit
	block := make(chan os.Signal, 1)
	signal.Notify(block, os.Interrupt)

	// Block until signal is received
	<-block
}
