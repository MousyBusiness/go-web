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
