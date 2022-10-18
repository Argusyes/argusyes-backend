package wsocket

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type ConnectHandler func(conn *Connect)
type MessageHandler func(conn *Connect, msg string)
type CloseHandler func(conn *Connect)
type ErrorHandler func(conn *Connect, err error)

type Connect struct {
	key     string
	conn    *websocket.Conn
	m       sync.Mutex
	manager *Manager
}

func (c *Connect) WriteMessage(data []byte) {
	c.m.Lock()
	defer c.m.Unlock()
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		delete(c.manager.websocketMap, c.key)
		c.manager.errorHandlerMutex.Lock()
		for _, h := range c.manager.errorHandler {
			h(c, err)
		}
		c.manager.errorHandlerMutex.Unlock()
	}
}

type Manager struct {
	websocketMapMutex   sync.Mutex
	websocketMap        map[string]*Connect
	connectHandlerMutex sync.Mutex
	connectHandler      []ConnectHandler
	messageHandlerMutex sync.Mutex
	messageHandler      []MessageHandler
	closeHandlerMutex   sync.Mutex
	closeHandler        []CloseHandler
	errorHandlerMutex   sync.Mutex
	errorHandler        []ErrorHandler
}

func newManager() (m *Manager) {
	return &Manager{
		websocketMap:   make(map[string]*Connect),
		connectHandler: make([]ConnectHandler, 0),
		messageHandler: make([]MessageHandler, 0),
		closeHandler:   make([]CloseHandler, 0),
		errorHandler:   make([]ErrorHandler, 0),
	}
}

var WsocketManager = newManager()

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func (m *Manager) HandleNewConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	id := uuid.New()
	key := fmt.Sprintf("%s:%s", conn.RemoteAddr().String(), id.String())
	if err != nil {
		log.Printf("websocket upgrade %s fail : %v", key, err)
		return
	}
	c := &Connect{
		key:     key,
		conn:    conn,
		manager: m,
	}
	m.websocketMapMutex.Lock()
	m.websocketMap[key] = c
	m.websocketMapMutex.Unlock()
	log.Printf("websocket connected from : %s", key)

	m.connectHandlerMutex.Lock()
	for _, h := range m.connectHandler {
		h(c)
	}
	m.connectHandlerMutex.Unlock()

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("websocket closed : %s fail %v", key, err)
		} else {
			log.Printf("websocket closed : %s", key)
		}
		delete(m.websocketMap, key)
		m.closeHandlerMutex.Lock()
		for _, h := range m.closeHandler {
			h(c)
		}
		m.closeHandlerMutex.Unlock()
	}()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("websocket %s error: %v", key, err)
			delete(m.websocketMap, key)
			m.errorHandlerMutex.Lock()
			for _, h := range m.errorHandler {
				h(c, err)
			}
			m.errorHandlerMutex.Unlock()
			break
		}
		switch mt {
		case websocket.TextMessage:
			message := string(message)
			log.Printf("websocket %s recv test: %s", key, message)
			m.messageHandlerMutex.Lock()
			for _, h := range m.messageHandler {
				h(c, message)
			}
			m.messageHandlerMutex.Unlock()
		}
	}
}

func (m *Manager) RegisterConnectHandler(handler ConnectHandler) {
	m.connectHandlerMutex.Lock()
	m.connectHandler = append(m.connectHandler, handler)
	m.connectHandlerMutex.Unlock()
}

func (m *Manager) RegisterMessageHandler(handler MessageHandler) {
	m.messageHandlerMutex.Lock()
	m.messageHandler = append(m.messageHandler, handler)
	m.messageHandlerMutex.Unlock()
}

func (m *Manager) RegisterCloseHandler(handler CloseHandler) {
	m.closeHandlerMutex.Lock()
	m.closeHandler = append(m.closeHandler, handler)
	m.closeHandlerMutex.Unlock()
}

func (m *Manager) RegisterErrorHandler(handler ErrorHandler) {
	m.errorHandlerMutex.Lock()
	m.errorHandler = append(m.errorHandler, handler)
	m.errorHandlerMutex.Unlock()
}
