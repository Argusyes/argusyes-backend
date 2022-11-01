module wsocket

go 1.19

replace mutexMap => ../mutexMap
require (
	mutexMap v0.0.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
)
