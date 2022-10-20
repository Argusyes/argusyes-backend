package main

import (
	"github.com/goccy/go-json"
	"log"
	"message"
	"ssh"
	"strings"
	"wsocket"
)

type WSRequest struct {
	Id     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type WSMonitorSSHRequest struct {
	Id     string `json:"id"`
	Method string `json:"method"`
	Params []struct {
		Port   int    `json:"port"`
		Host   string `json:"host"`
		User   string `json:"user"`
		Passwd string `json:"passwd"`
	} `json:"params"`
}

type WSUnMonitorSSHRequest struct {
	Id     string `json:"id"`
	Method string `json:"method"`
	Params []struct {
		Port int    `json:"port"`
		Host string `json:"host"`
		User string `json:"user"`
	} `json:"params"`
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
	case "connect_ssh":
		wsMonitorSSHRequest := &WSMonitorSSHRequest{}
		err := json.Unmarshal(msg, wsMonitorSSHRequest)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		for _, p := range wsMonitorSSHRequest.Params {
			s := ssh.SSHManager.GetSSH(p.Port, p.Host, p.User, p.Passwd)
			s.RegisterCPUInfoListener(conn.Key, func(msg message.CPUInfoMessage) {
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					log.Println("json:", err)
				}
				conn.WriteMessage(msgBytes)
			})
		}
	case "disconnect_ssh":
		wsUnMonitorSSHRequest := &WSUnMonitorSSHRequest{}
		err := json.Unmarshal(msg, wsUnMonitorSSHRequest)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		sshKeys := make([]string, 0)
		for _, p := range wsUnMonitorSSHRequest.Params {
			sshKeys = append(sshKeys, ssh.GeneralKey(p.Port, p.Host, p.User))
		}
		ssh.SSHManager.RemoveSSHCPUInfoListener(sshKeys, conn.Key)
	}
}

func handleResponse(conn *wsocket.Connect, msg []byte) {
	log.Printf("recv response : %s", string(msg))
}
