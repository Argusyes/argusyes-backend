package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml"
	"log"
	"mongoDB"
	"net/http"
	"ssh"
	"wsocket"
)

var valid *validator.Validate

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	valid = validator.New()
}

func main() {
	defer mongoDB.Client.Close()

	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}

	ip := conf.Get("server.IP").(string)
	port := conf.Get("server.Port").(int64)
	allowOrigin := conf.Get("server.AllowOrigin").(string)
	addr := fmt.Sprintf("%s:%d", ip, port)

	router := gin.New()
	router.Use(ginMiddleware(allowOrigin))
	router.GET("/monitor", monitorHandler)
	router.POST("/user/addSSH", addUserSSHHandler)

	wsocket.WsocketManager.RegisterMessageHandler(messageRouter)
	wsocket.WsocketManager.RegisterCloseHandler(func(conn *wsocket.Connect) {
		ssh.SSHManager.ClearListener(conn.Key)
	})
	wsocket.WsocketManager.RegisterErrorHandler(func(conn *wsocket.Connect, err error) {
		ssh.SSHManager.ClearListener(conn.Key)
	})

	if err := router.Run(addr); err != nil {
		log.Fatal("failed run app: ", err)
	}
}

func monitorHandler(c *gin.Context) {
	wsocket.WsocketManager.HandleNewConnect(c.Writer, c.Request)
}

func ginMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}
