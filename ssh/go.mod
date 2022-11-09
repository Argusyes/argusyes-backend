module ssh

go 1.18

replace (
	logger => ../logger
	mutexMap => ../mutexMap
)

require (
	github.com/deckarep/golang-set/v2 v2.1.0
	github.com/gorilla/websocket v1.5.0
	github.com/pkg/sftp v1.13.5
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
	logger v0.0.0
	mutexMap v0.0.0

)

require (
	github.com/kr/fs v0.1.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
