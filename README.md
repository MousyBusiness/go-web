# HTTP and WebSockets helper with simple test mocking

### Quick Start HTTP

##### Simple GET
```
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

```
> Content-Type will default to json

##### Simple POST
```
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

	code, bytes, err := web.Post(url, time.Second*10, []byte("{}"))
	
	log.Println(code, string(bytes), err)
}
```

##### Authenticated POST
```
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
```
> [Setting up firebase authentication](https://www.youtube.com/watch?v=A2TqeQRQHL0&feature=youtu.be)
----

### Quick Start WebSockets

##### Authenticated Websocket Server
```
package main

import (
	"context"
	fbauth "firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/mousybusiness/go-web/ws/server"
	"github.com/mousybusiness/googlecloudgo/pkg/auth"
	errs "github.com/pkg/errors"
	"io"
	"log"
	"net/http"
)

type Cleanable struct {
	conn io.ReadWriteCloser
}

func (c Cleanable) CleanUp(uid string) error {
	delete(server.Connections, uid) // remove connections from local list
	_ = c.conn.Close()
	return nil
}

func (c Cleanable) GetConnection() io.ReadWriteCloser {
	return c.conn
}

func websocketHandler(c *gin.Context) {
	// get firebase IdToken from Gin context
	t, ok := c.Get(auth.FirebaseContextVal)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	token := t.(*fbauth.Token)

	if _, ok := server.Connections[token.UID]; !ok {
		// upgrade to websocket
		conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
		if err != nil {
			log.Println(errs.Wrap(err, "couldn't upgrade websocket"))
			c.AbortWithStatusJSON(http.StatusUpgradeRequired, gin.H{"error": http.StatusText(http.StatusUpgradeRequired)})
			return
		}

		cc := server.NewConnection(token.UID, Cleanable{
			conn: conn,
		})

		// read stuff
		cc.Read(context.TODO(), nil)

	} else {
		log.Println("conflict, connection already exists,", server.Connections[token.UID])
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "connection already exists"})
	}
}

func main() {
	// setup GIN
	r := gin.Default()

	// setup Firebase
	var err error

	// depends on os.Getenv("FIREBASE_CONFIG_FILE")
	fbclient, err := auth.InitAuth()
	if err != nil {
		log.Fatalln("failed to init firebase auth", err)
	}

	// use JWT auth middleware
	authed := r.Group("")
	authed.Use(auth.AuthJWT(fbclient))

	// websocket route
	authed.GET("/connect", websocketHandler)

	// start API
	log.Fatalln(r.Run(":80"))
}
  
 ```
> gobwas has the potential to [create millions](http://goroutines.com/10m) of simultaneous websocket connections on a single server if you plan on implementing your own notification system


##### Authenticated Websocket Client
```
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

	conn, _ := client.NewConnection(websocket.DefaultDialer, "my-connection-name", "http://myapp.com", "/signal", os.Getenv("TOKEN"))

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

```
