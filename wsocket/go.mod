module wsocket

go 1.19

replace (
	logger => ../logger
	mutexMap => ../mutexMap
)

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	logger v0.0.0-00010101000000-000000000000
	mutexMap v0.0.0
)

require (
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
