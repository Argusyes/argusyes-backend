package main

import (
	"fmt"
	socketIO "github.com/googollee/go-socket.io"
	"github.com/pelletier/go-toml"
	"log"
	"message"
	"net/http"
	"ssh"
)

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile)
}

func main() {
	defer ssh.Client.Close()

	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}

	ip := conf.Get("server.IP").(string)
	port := conf.Get("server.Port").(int64)
	addr := fmt.Sprintf("%s:%d", ip, port)

	server := socketIO.NewServer(nil)

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatalf("socket io server start error : %v", err)
		}
	}()

	defer func(server *socketIO.Server) {
		err := server.Close()
		if err != nil {
			log.Fatalf("socket io server close error : %v", err)
		}
	}(server)

	server.OnConnect("/", func(s socketIO.Conn) error {
		s.SetContext("")
		s.Join("notification")
		return nil
	})

	ssh.Client.RegisterCPUInfoListener(func(m message.CPUInfoMessage) {
		server.BroadcastToRoom("", "notification", "event:name", m)
	})
	ssh.Client.StartAllMonitor()
	http.Handle("/monitor/", server)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("http server error : %v", err)
	}
}
