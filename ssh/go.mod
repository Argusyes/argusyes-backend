module ssh

go 1.18

replace mutexMap => ../mutexMap

require (
	mutexMap v0.0.0
	github.com/deckarep/golang-set/v2 v2.1.0
	github.com/gorilla/websocket v1.5.0
	github.com/pkg/sftp v1.13.5
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
	
)

require (
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)
