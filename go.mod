module argus

go 1.18

replace (
	message => ./message
	ssh => ./ssh
)

require ssh v0.0.0

require (
	github.com/googollee/go-socket.io v1.6.2
	github.com/pelletier/go-toml v1.9.5
)

require (
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	message v0.0.0 // indirect
)
