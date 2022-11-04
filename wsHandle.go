package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"log"
	"regexp"
	"ssh"
	"strings"
	"sync"
	"wsocket"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RequestHead struct {
	Id     *string `json:"id"`
	Method string  `json:"method" validate:"required"`
}

type WSRequest struct {
	RequestHead
	Params interface{} `json:"params"`
}

type WSMonitorSSHRequest struct {
	RequestHead
	Params []struct {
		Port   int    `json:"port"`
		Host   string `json:"host" validate:"required"`
		User   string `json:"user" validate:"required"`
		Passwd string `json:"passwd" validate:"required"`
	} `json:"params" validate:"required,dive"`
}

type WSUnMonitorSSHRequest struct {
	RequestHead
	Params []struct {
		Port int    `json:"port" validate:"required"`
		Host string `json:"host" validate:"required"`
		User string `json:"user" validate:"required"`
	} `json:"params" validate:"required,dive"`
}

type WSStartSSHRequest struct {
	RequestHead
	Params []struct {
		Port   int    `json:"port" validate:"required"`
		Host   string `json:"host" validate:"required"`
		User   string `json:"user" validate:"required"`
		Passwd string `json:"passwd" validate:"required"`
	} `json:"params" validate:"required,dive"`
}

type WSNotificationRequest struct {
	RequestHead
	Params []struct {
		Event   string      `json:"event" validate:"required"`
		Message interface{} `json:"message" validate:"required"`
	} `json:"params" validate:"required,dive"`
}

type ResponseHead struct {
	Id    string         `json:"id" validate:"required"`
	Error *ResponseError `json:"error"`
}

type ResponseError struct {
	Code    int    `json:"code" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type WSResponse struct {
	ResponseHead
	Result interface{} `json:"result"`
}

type WSMonitorSSHResponse struct {
	ResponseHead
	Result []WSMonitorSSHResponseResult `json:"result" validate:"dive"`
}

type WSMonitorSSHResponseResult struct {
	Port    int            `json:"port" validate:"required"`
	Host    string         `json:"host" validate:"required"`
	User    string         `json:"user" validate:"required"`
	Monitor bool           `json:"monitor"`
	Error   *ResponseError `json:"error"`
}

type WSUnMonitorSSHResponse struct {
	ResponseHead
	Result []WSUnMonitorSSHResponseResult `json:"result" validate:"dive"`
}

type WSUnMonitorSSHResponseResult struct {
	Port      int            `json:"port" validate:"required"`
	Host      string         `json:"host" validate:"required"`
	User      string         `json:"user" validate:"required"`
	UnMonitor bool           `json:"unMonitor"`
	Error     *ResponseError `json:"error"`
}

type WSStartSSHResponse struct {
	ResponseHead
	Result []bool `json:"result" validate:"dive"`
}

func listenerTemplate[M any](conn *wsocket.Connect, event string) func(m M) {
	return func(m M) {
		request := &WSNotificationRequest{
			RequestHead: RequestHead{
				Id:     nil,
				Method: "ssh.notification",
			},
			Params: []struct {
				Event   string      `json:"event" validate:"required"`
				Message interface{} `json:"message" validate:"required"`
			}{{
				Event:   event,
				Message: m,
			}},
		}

		requestBytes, err := json.Marshal(request)
		if err != nil {
			log.Fatalf("json parse fail : %v", err)
		}
		conn.WriteMessage(requestBytes)
	}
}

func getSSHListener(conn *wsocket.Connect) ssh.AllListener {

	return ssh.AllListener{
		CPUInfoListener:           listenerTemplate[ssh.CPUInfoMessage](conn, "cpuInfo"),
		CPUPerformanceListener:    listenerTemplate[ssh.CPUPerformanceMessage](conn, "cpuPerformance"),
		MemoryPerformanceListener: listenerTemplate[ssh.MemoryPerformanceMessage](conn, "memoryPerformance"),
		UptimeListener:            listenerTemplate[ssh.UptimeMessage](conn, "uptime"),
		LoadavgListener:           listenerTemplate[ssh.LoadavgMessage](conn, "loadavg"),
		NetDevListener:            listenerTemplate[ssh.NetDevMessage](conn, "netDev"),
		NetStatListener:           listenerTemplate[ssh.NetStatMessage](conn, "netStat"),
		TempListener:              listenerTemplate[ssh.TempMessage](conn, "temp"),
		DiskListener:              listenerTemplate[ssh.DiskMessage](conn, "disk"),
		ProcessListener:           listenerTemplate[ssh.ProcessMessage](conn, "process"),
	}
}

func messageJsonParseHelper(id string, conn *wsocket.Connect, msg []byte, v interface{}) bool {
	err := json.Unmarshal(msg, v)
	if err != nil {
		errText := fmt.Sprintf("Json parse fail : %v", err)
		log.Printf(errText)
		wsResponse := &WSResponse{
			ResponseHead: ResponseHead{
				Id: id,
				Error: &ResponseError{
					Code:    400,
					Message: errText,
				},
			},
			Result: nil,
		}
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsResponse); ok {
			conn.WriteMessage(wsResponseBytes)
		}
		return false
	} else if err := valid.Struct(v); err != nil {
		errText := fmt.Sprintf("message validate fail : %v", err)
		log.Printf(errText)
		wsResponse := &WSResponse{
			ResponseHead: ResponseHead{
				Id: id,
				Error: &ResponseError{
					Code:    400,
					Message: errText,
				},
			},
			Result: nil,
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
	log.Printf("%s handle message %s", conn.Key, string(msg))
	ms := string(msg)
	if strings.Contains(ms, "method") {
		handleRequest(conn, msg)
	} else if strings.Contains(ms, "result") {
		handleResponse(conn, msg)
	} else {
		log.Printf("unknown msg : %s", ms)
	}
}

func handleRequestCommon(msg []byte, conn *wsocket.Connect) (id, method string, ok bool) {
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
			return "", "", false
		}
		wsResponse := &WSResponse{
			ResponseHead: ResponseHead{
				Id: string(idRegResults[0][1]),
				Error: &ResponseError{
					Code:    400,
					Message: errText,
				},
			},
			Result: nil,
		}
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsResponse); ok {
			conn.WriteMessage(wsResponseBytes)
		}
		return "", "", false
	}
	return *wsRequest.Id, wsRequest.Method, true
}

func handleRequest(conn *wsocket.Connect, msg []byte) {
	id, method, ok := handleRequestCommon(msg, conn)
	if !ok {
		return
	}
	switch method {
	case "ssh.startRoughMonitor":
		log.Printf("%s handle ssh.startMonitior", conn.Key)
		wsMonitorSSHRequest := &WSMonitorSSHRequest{}
		if ok := messageJsonParseHelper(id, conn, msg, wsMonitorSSHRequest); !ok {
			return
		}
		wsMonitorSSHResponse := &WSMonitorSSHResponse{
			ResponseHead: ResponseHead{
				Id:    *wsMonitorSSHRequest.Id,
				Error: nil,
			},
			Result: make([]WSMonitorSSHResponseResult, 0),
		}
		m := sync.Mutex{}
		wg := sync.WaitGroup{}

		for _, p := range wsMonitorSSHRequest.Params {
			wg.Add(1)
			go func(port int, host string, user string, passwd string) {
				err := ssh.M.RegisterRoughListener(port, host, user, passwd, conn.Key, listenerTemplate[ssh.RoughMessage](conn, "rough"))
				result := WSMonitorSSHResponseResult{
					Port: port,
					Host: host,
					User: user,
				}
				if err == nil {
					result.Monitor = true
					result.Error = nil
				} else {
					errText := err.Error()
					result.Monitor = false
					result.Error = &ResponseError{
						Code:    400,
						Message: errText,
					}
				}
				m.Lock()
				wsMonitorSSHResponse.Result = append(wsMonitorSSHResponse.Result, result)
				m.Unlock()
				wg.Done()
				log.Printf("rough done %d %s %s", port, host, user)
			}(p.Port, p.Host, p.User, p.Passwd)
		}
		wg.Wait()
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsMonitorSSHResponse); ok {
			conn.WriteMessage(wsResponseBytes)
			log.Printf("send to wsocket %s", string(wsResponseBytes))
		}

	case "ssh.stopRoughMonitor":
		wsUnMonitorSSHRequest := &WSUnMonitorSSHRequest{}
		if ok := messageJsonParseHelper(id, conn, msg, wsUnMonitorSSHRequest); !ok {
			return
		}
		wsUnMonitorSSHResponse := &WSUnMonitorSSHResponse{
			ResponseHead: ResponseHead{
				Id:    *wsUnMonitorSSHRequest.Id,
				Error: nil,
			},
			Result: make([]WSUnMonitorSSHResponseResult, 0),
		}
		for _, p := range wsUnMonitorSSHRequest.Params {
			ssh.M.RemoveRoughListener(p.Port, p.Host, p.User, conn.Key)
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
			log.Printf("send to wsocket %s", string(wsResponseBytes))
		}
	case "ssh.startMonitor":
		log.Printf("%s handle ssh.startMonitior", conn.Key)
		wsMonitorSSHRequest := &WSMonitorSSHRequest{}
		if ok := messageJsonParseHelper(id, conn, msg, wsMonitorSSHRequest); !ok {
			return
		}
		wsMonitorSSHResponse := &WSMonitorSSHResponse{
			ResponseHead: ResponseHead{
				Id:    *wsMonitorSSHRequest.Id,
				Error: nil,
			},
			Result: make([]WSMonitorSSHResponseResult, 0),
		}
		m := sync.Mutex{}
		wg := sync.WaitGroup{}

		for _, p := range wsMonitorSSHRequest.Params {
			wg.Add(1)
			go func(port int, host string, user string, passwd string) {
				err := ssh.M.RegisterSSHListener(port, host, user, passwd, conn.Key, getSSHListener(conn))
				result := WSMonitorSSHResponseResult{
					Port: port,
					Host: host,
					User: user,
				}
				if err == nil {
					result.Monitor = true
					result.Error = nil
				} else {
					errText := err.Error()
					result.Monitor = false
					result.Error = &ResponseError{
						Code:    400,
						Message: errText,
					}
				}
				m.Lock()
				wsMonitorSSHResponse.Result = append(wsMonitorSSHResponse.Result, result)
				m.Unlock()
				wg.Done()
				log.Printf("done %d %s %s", port, host, user)
			}(p.Port, p.Host, p.User, p.Passwd)
		}
		wg.Wait()
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsMonitorSSHResponse); ok {
			conn.WriteMessage(wsResponseBytes)
			log.Printf("send to wsocket %s", string(wsResponseBytes))
		}
	case "ssh.stopMonitor":
		wsUnMonitorSSHRequest := &WSUnMonitorSSHRequest{}
		if ok := messageJsonParseHelper(id, conn, msg, wsUnMonitorSSHRequest); !ok {
			return
		}
		wsUnMonitorSSHResponse := &WSUnMonitorSSHResponse{
			ResponseHead: ResponseHead{
				Id:    *wsUnMonitorSSHRequest.Id,
				Error: nil,
			},
			Result: make([]WSUnMonitorSSHResponseResult, 0),
		}
		for _, p := range wsUnMonitorSSHRequest.Params {
			ssh.M.RemoveSSHListener(p.Port, p.Host, p.User, conn.Key)
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
			log.Printf("send to wsocket %s", string(wsResponseBytes))
		}
	}
}

func handleResponse(conn *wsocket.Connect, msg []byte) {
	log.Printf("recv response : %s", string(msg))
}
