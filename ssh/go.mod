module ssh

go 1.18

replace message => ../message

require message v0.0.0

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/pelletier/go-toml v1.9.5
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
)

require golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
