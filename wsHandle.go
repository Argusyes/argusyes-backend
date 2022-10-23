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

type WSNotificationRequest struct {
	Id     *string `json:"id"`
	Method string  `json:"method"`
	Params []struct {
		Event   string                 `json:"event"`
		Message message.CPUInfoMessage `json:"message"`
	} `json:"params"`
}

type WSResponse struct {
	Id     string      `json:"id"`
	Error  *string     `json:"error"`
	Result interface{} `json:"result"`
}

type WSMonitorSSHResponse struct {
	Id     string                       `json:"id"`
	Error  *string                      `json:"error"`
	Result []WSMonitorSSHResponseResult `json:"result"`
}

type WSMonitorSSHResponseResult struct {
	Port    int     `json:"port"`
	Host    string  `json:"host"`
	User    string  `json:"user"`
	Monitor bool    `json:"monitor"`
	Error   *string `json:"error"`
}

type WSUnMonitorSSHResponse struct {
	Id     string                         `json:"id"`
	Error  *string                        `json:"error"`
	Result []WSUnMonitorSSHResponseResult `json:"result"`
}

type WSUnMonitorSSHResponseResult struct {
	Port      int     `json:"port"`
	Host      string  `json:"host"`
	User      string  `json:"user"`
	UnMonitor bool    `json:"un_monitor"`
	Error     *string `json:"error"`
}

func getSSHListener(conn *wsocket.Connect) *ssh.Listener {
	return &ssh.Listener{
		CPUInfoListener: func(m message.CPUInfoMessage) {
			request := &WSNotificationRequest{
				Id:     nil,
				Method: "ssh.notification",
				Params: []struct {
					Event   string                 `json:"event"`
					Message message.CPUInfoMessage `json:"message"`
				}{{
					Event:   "cpu_info",
					Message: m,
				}},
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
		errText := fmt.Sprintf("Json parse fail : %v", err)
		log.Printf(errText)
		wsResponse := &WSResponse{
			Id:     *wsRequest.Id,
			Result: nil,
			Error:  &errText,
		}
		wsResponseBytes, err := json.Marshal(wsResponse)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		conn.WriteMessage(wsResponseBytes)
		return
	}
	switch wsRequest.Method {
	case "ssh.start_monitor":
		wsMonitorSSHRequest := &WSMonitorSSHRequest{}
		err := json.Unmarshal(msg, wsMonitorSSHRequest)
		if err != nil {
			errText := fmt.Sprintf("Json parse fail : %v", err)
			log.Printf(errText)
			wsResponse := &WSResponse{
				Id:     *wsRequest.Id,
				Result: nil,
				Error:  &errText,
			}
			wsResponseBytes, err := json.Marshal(wsResponse)
			if err != nil {
				log.Printf("Json parse fail : %v", err)
			}
			conn.WriteMessage(wsResponseBytes)
			return
		}
		wsMonitorSSHResponse := &WSMonitorSSHResponse{
			Id:     *wsMonitorSSHRequest.Id,
			Error:  nil,
			Result: make([]WSMonitorSSHResponseResult, 0),
		}
		for _, p := range wsMonitorSSHRequest.Params {
			err := ssh.SSHManager.RegisterAllMonitorListener(p.Port, p.Host, p.User, p.Passwd, conn.Key, getSSHListener(conn))
			result := WSMonitorSSHResponseResult{
				Port: p.Port,
				Host: p.Host,
				User: p.User,
			}
			if err == nil {
				result.Monitor = true
				result.Error = nil
			} else {
				errText := err.Error()
				result.Monitor = false
				result.Error = &errText
			}
			wsMonitorSSHResponse.Result = append(wsMonitorSSHResponse.Result, result)
		}
		wsResponseBytes, err := json.Marshal(wsMonitorSSHResponse)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		conn.WriteMessage(wsResponseBytes)

	case "ssh.stop_monitor":

		wsUnMonitorSSHRequest := &WSUnMonitorSSHRequest{}
		err := json.Unmarshal(msg, wsUnMonitorSSHRequest)
		if err != nil {
			eString := fmt.Sprintf("Json parse fail : %v", err)
			log.Printf(eString)
			wsResponse := &WSResponse{
				Id:     *wsRequest.Id,
				Result: nil,
				Error:  &eString,
			}
			wsResponseBytes, err := json.Marshal(wsResponse)
			if err != nil {
				log.Printf("Json parse fail : %v", err)
			}
			conn.WriteMessage(wsResponseBytes)
			return
		}
		wsUnMonitorSSHResponse := &WSUnMonitorSSHResponse{
			Id:     *wsUnMonitorSSHRequest.Id,
			Error:  nil,
			Result: make([]WSUnMonitorSSHResponseResult, 0),
		}
		for _, p := range wsUnMonitorSSHRequest.Params {
			ssh.SSHManager.RemoveSSHListener(p.Port, p.Host, p.User, conn.Key)
			result := WSUnMonitorSSHResponseResult{
				Port:      p.Port,
				Host:      p.Host,
				User:      p.User,
				UnMonitor: true,
				Error:     nil,
			}
			wsUnMonitorSSHResponse.Result = append(wsUnMonitorSSHResponse.Result, result)
		}
		wsResponseBytes, err := json.Marshal(wsUnMonitorSSHResponse)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		conn.WriteMessage(wsResponseBytes)
	}
}

func handleResponse(conn *wsocket.Connect, msg []byte) {
	log.Printf("recv response : %s", string(msg))
}
