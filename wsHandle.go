package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"log"
	"message"
	"regexp"
	"ssh"
	"strings"
	"wsocket"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
		Host   string `json:"host" validate:"ip_addr"`
		User   string `json:"user"`
		Passwd string `json:"passwd"`
	} `json:"params" validate:"required,dive"`
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

func messageJsonParseHelper(id string, conn *wsocket.Connect, msg []byte, v interface{}) bool {
	err := json.Unmarshal(msg, v)
	if err != nil {
		errText := fmt.Sprintf("Json parse fail : %v", err)
		log.Printf(errText)
		wsResponse := &WSResponse{
			Id:     id,
			Result: nil,
			Error:  &errText,
		}
		wsResponseBytes, err := json.Marshal(wsResponse)
		if err != nil {
			log.Printf("Json parse fail : %v", err)
		}
		conn.WriteMessage(wsResponseBytes)
		return false
	} else if err := valid.Struct(v); err != nil {
		errText := fmt.Sprintf("message validate fail : %v", err)
		log.Printf(errText)
		wsResponse := &WSResponse{
			Id:     id,
			Result: nil,
			Error:  &errText,
		}
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsResponse); ok {
			conn.WriteMessage(wsResponseBytes)
		}
		return false
	}
	return true
}

func messageJsonStringifyHelper(v interface{}) ([]byte, bool) {
	bytes, err := json.Marshal(v)
	if err != nil {
		errText := fmt.Sprintf("Json parse fail : %v", err)
		log.Printf(errText)
		return nil, false
	} else if err := valid.Struct(v); err != nil {
		errText := fmt.Sprintf("message validate fail : %v", err)
		log.Printf(errText)
		return nil, false
	}
	return bytes, true
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
		idReg := regexp.MustCompile(`"id":"([^)]+)"`)
		if idReg == nil {
			log.Fatalf("regexp parse fail : id")
		}
		idRegResults := idReg.FindAllSubmatch(msg, -1)
		if idRegResults == nil {
			log.Printf("parse id fail")
			return
		}
		wsResponse := &WSResponse{
			Id:     string(idRegResults[0][1]),
			Result: nil,
			Error:  &errText,
		}
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsResponse); ok {
			conn.WriteMessage(wsResponseBytes)
		}
		return
	}
	switch wsRequest.Method {
	case "ssh.start_monitor":
		wsMonitorSSHRequest := &WSMonitorSSHRequest{}
		if ok := messageJsonParseHelper(*wsRequest.Id, conn, msg, wsMonitorSSHRequest); !ok {
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
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsMonitorSSHResponse); ok {
			conn.WriteMessage(wsResponseBytes)
		}
	case "ssh.stop_monitor":

		wsUnMonitorSSHRequest := &WSUnMonitorSSHRequest{}
		if ok := messageJsonParseHelper(*wsRequest.Id, conn, msg, wsUnMonitorSSHRequest); !ok {
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
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsUnMonitorSSHResponse); ok {
			conn.WriteMessage(wsResponseBytes)
		}
	}
}

func handleResponse(conn *wsocket.Connect, msg []byte) {
	log.Printf("recv response : %s", string(msg))
}
