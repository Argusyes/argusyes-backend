package main

import (
	"fmt"
	"github.com/goccy/go-json"
	"log"
	"message"
	"ssh"
	"strings"
	"wsocket"
)

type WSRequest struct {
	Id     *string     `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type WSMonitorSSHRequest struct {
	Id     *string `json:"id"`
	Method string  `json:"method"`
	Params []struct {
		Port   int    `json:"port"`
		Host   string `json:"host"`
		User   string `json:"user"`
		Passwd string `json:"passwd"`
	} `json:"params"`
}

type WSUnMonitorSSHRequest struct {
	Id     *string `json:"id"`
	Method string  `json:"method"`
	Params []struct {
		Port int    `json:"port"`
		Host string `json:"host"`
		User string `json:"user"`
	} `json:"params"`
}

type WSResponse struct {
	Id     string  `json:"id"`
	Result *string `json:"result"`
	Error  *string `json:"error"`
}

func getSSHListener(conn *wsocket.Connect) *ssh.Listener {
	return &ssh.Listener{
		CPUInfoListener: func(m message.CPUInfoMessage) {
			param := [...]message.CPUInfoMessage{m}
			request := &WSRequest{
				Id:     nil,
				Method: "ssh.notification",
				Params: param,
			}

			requestBytes, err := json.Marshal(request)
			if err != nil {
				log.Fatalf("json parse fail : %v", err)
			}
			conn.WriteMessage(requestBytes)
		},
	}
}

func messageRouter(conn *wsocket.Connect, msg []byte) {
	ms := string(msg)
	if strings.Contains(ms, "method") {
		handleRequest(conn, msg)
	} else if strings.Contains(ms, "result") {
		handleResponse(conn, msg)
	} else {
		log.Printf("unknown msg : %s", ms)
	}
}

func handleRequest(conn *wsocket.Connect, msg []byte) {
	wsRequest := &WSRequest{}
	err := json.Unmarshal(msg, wsRequest)
	if err != nil {
		log.Printf("Json parse fail : %v", err)
	}
	switch wsRequest.Method {
	case "ssh.start_monitor":
		wsResponse := &WSResponse{
			Id:     *wsRequest.Id,
			Result: nil,
			Error:  nil,
		}
		wsMonitorSSHRequest := &WSMonitorSSHRequest{}
		err := json.Unmarshal(msg, wsMonitorSSHRequest)
		if err != nil {
			eString := fmt.Sprintf("Json parse fail : %v", err)
			wsResponse.Error = &eString
			log.Printf("Json parse fail : %v", err)
		}
		for _, p := range wsMonitorSSHRequest.Params {
			ssh.SSHManager.RegisterAllMonitorListener(p.Port, p.Host, p.User, p.Passwd, conn.Key, getSSHListener(conn))
		}

		wsResponseBytes, err := json.Marshal(wsResponse)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		conn.WriteMessage(wsResponseBytes)

	case "ssh.stop_monitor":
		wsResponse := &WSResponse{
			Id:     *wsRequest.Id,
			Result: nil,
			Error:  nil,
		}
		wsUnMonitorSSHRequest := &WSUnMonitorSSHRequest{}
		err := json.Unmarshal(msg, wsUnMonitorSSHRequest)
		if err != nil {
			eString := fmt.Sprintf("Json parse fail : %v", err)
			wsResponse.Error = &eString
			log.Printf("Json parse fail : %v", err)
		}
		for _, p := range wsUnMonitorSSHRequest.Params {
			ssh.SSHManager.RemoveSSHListener(p.Port, p.Host, p.User, conn.Key)
		}
		wsResponseBytes, err := json.Marshal(wsResponse)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		conn.WriteMessage(wsResponseBytes)
	}
}

func handleResponse(conn *wsocket.Connect, msg []byte) {
	log.Printf("recv response : %s", string(msg))
}
