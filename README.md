# HTTP and WebSockets helper

### Quick Start HTTP

##### Simple GET
```
url := "http://metadata.google.internal/computeMetadata/v1/project/project-id"

code, bytes, err := web.Get(url, time.Second*2, web.KV{"Metadata-Flavor", "Google"})
  
```
> Content-Type will default to json

##### Simple POST
```
url := fmt.Sprintf("http://%s/test", yourHost) 

code, bytes, err := web.Post(url, time.Second*10, msg)
  
```

##### Authenticated POST
```
if os.Getenv("TOKEN") == "" {
  log.Fatalln("require token environment variable")
}

url := fmt.Sprintf("http://%s/test", yourHost) 

// by using APost (Authenticated Post)  KV{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TOKEN"))} will be added as a header
code, bytes, err := web.APost(url, time.Second*10, msg)
  
```

----

### Quick Start WebSockets

##### Authenticated Websocket Server
```
import (
	"context"
	"encoding/json"
	"errors"
	fbauth "firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/mousybusiness/googlecloudgo/pkg/auth"
	"github.com/mousybusiness/go-web/ws/server"
	"io"
	"log"
	"net/http"
)

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
	fbclient, err = auth.InitAuth()
	if err != nil {
		log.Fatalln("failed to init firebase auth", err)
	}

  // use JWT auth middleware
  authed := r.Group("")
  authed.Use(auth.AuthJWT(fbclient))

  // websocket route
  authed.GET("/connect", websocketHandler)
  
  // start API
  log.Fatalln(setupRouter().Run(":80"))
}
  
 ```
> gobwas has the potential to [create millions](http://goroutines.com/10m) of simultaneous websocket connections on a single server if you plan on implementing your own notification system


##### Authenticated Websocket Client
```
import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/mousybusiness/go-web/ws/client"
	errs "github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"os"
  "os/signal"
)

func handleMsg(msg []byte) error {
  log.Println(string(msg))
  return nil
}

func main() {
    ctx := context.Background()

    conn, err := client.NewConnection(websocket.DefaultDialer, "my-connection-name", "http://myapp.com", "/signal", os.Getenv("TOKEN"))

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
