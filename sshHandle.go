package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"logger"
	"net/http"
	"regexp"
	"ssh"
	"sync"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func handleNewSSHConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	id := uuid.New()
	key := fmt.Sprintf("%s:%s", conn.RemoteAddr().String(), id.String())
	if err != nil {
		logger.L.Debugf("websocket upgrade %s fail : %v", key, err)
		return
	}
	logger.L.Debugf("websocket connected from : %s", key)

	mt, message, err := conn.ReadMessage()
	if err != nil {
		logger.L.Debugf("websocket %s error: %v", key, err)
		return
	}
	switch mt {
	case websocket.TextMessage:
		logger.L.Debugf("websocket %s recv : %s", key, message)
		wsRequest := &WSRequest{}
		err := json.Unmarshal(message, wsRequest)
		if err != nil || wsRequest.Method != "ssh.startSSH" {
			errText := fmt.Sprintf("Json parse fail : %v", err)
			logger.L.Debugf(errText)
			idReg := regexp.MustCompile(`"id":"([^)]+)"`)
			if idReg == nil {
				logger.L.Fatalf("regexp parse fail : id")
			}
			idRegResults := idReg.FindAllSubmatch(message, -1)
			if idRegResults == nil {
				logger.L.Debugf("parse id fail")
				return
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
				_ = conn.WriteMessage(websocket.TextMessage, wsResponseBytes)
			}
			return
		}

		id := *wsRequest.Id
		wsStartSSHRequest := &WSStartSSHRequest{}
		err = json.Unmarshal(message, wsStartSSHRequest)
		if err != nil {
			errText := fmt.Sprintf("Json parse fail : %v", err)
			logger.L.Debugf(errText)
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
				_ = conn.WriteMessage(websocket.TextMessage, wsResponseBytes)
			}
			return
		} else if err := valid.Struct(wsStartSSHRequest); err != nil {
			errText := fmt.Sprintf("message validate fail : %v", err)
			logger.L.Debugf(errText)
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
				_ = conn.WriteMessage(websocket.TextMessage, wsResponseBytes)
			}
			return
		}
		wsStartSSHResponse := &WSStartSSHResponse{
			ResponseHead: ResponseHead{
				Id:    *wsStartSSHRequest.Id,
				Error: nil,
			},
			Result: make([]bool, 0),
		}
		m := &sync.Mutex{}
		res, err := ssh.M.NewSSHClientWithConn(wsStartSSHRequest.Params[0].Port, wsStartSSHRequest.Params[0].Host, wsStartSSHRequest.Params[0].User, wsStartSSHRequest.Params[0].Passwd, conn, m)
		if !res || err != nil {
			wsStartSSHResponse.Error = &ResponseError{
				Code:    400,
				Message: err.Error(),
			}
		} else {
			wsStartSSHResponse.Result = append(wsStartSSHResponse.Result, res)
		}
		if wsResponseBytes, ok := messageJsonStringifyHelper(wsStartSSHResponse); ok {
			m.Lock()
			_ = conn.WriteMessage(websocket.TextMessage, wsResponseBytes)
			m.Unlock()
		}
	}
}
