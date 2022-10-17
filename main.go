package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"github.com/pelletier/go-toml"
	"log"
	"message"
	"net/http"
	"ssh"
	"sync"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var upgrader = websocket.Upgrader{}
var mutex sync.Mutex
var m = make(map[string]*websocket.Conn)
var index = 0

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	key := fmt.Sprintf("%s:%d", c.RemoteAddr().String(), index)
	index++
	mutex.Lock()
	m[key] = c
	mutex.Unlock()
	log.Print("connect" + key)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(m, key)
			mutex.Unlock()
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		mutex.Lock()
		for _, v := range m {
			err = v.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
		mutex.Unlock()
	}
}

func main() {
	defer ssh.Client.Close()

	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}

	ip := conf.Get("server.IP").(string)
	port := conf.Get("server.Port").(int64)
	allowOrigin := conf.Get("server.AllowOrigin").(string)
	addr := fmt.Sprintf("%s:%d", ip, port)

	ssh.Client.RegisterCPUInfoListener(func(msg message.CPUInfoMessage) {
		mutex.Lock()
		for _, v := range m {
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Println("json:", err)
			}
			err = v.WriteMessage(1, msgBytes)
			if err != nil {
				log.Println("write:", err)
			}
		}
		mutex.Unlock()
	})
	ssh.Client.StartAllMonitor()

	router := gin.New()
	router.Use(GinMiddleware(allowOrigin))
	router.GET("/ws", func(c *gin.Context) {
		echo(c.Writer, c.Request)
	})

	if err := router.Run(addr); err != nil {
		log.Fatal("failed run app: ", err)
	}
}

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
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
