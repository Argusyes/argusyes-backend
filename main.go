package main

import (
	"fmt"
	mapSet "github.com/deckarep/golang-set/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml"
	"log"
	"mongoDB"
	"net/http"
	"net/url"
	"ssh"
	"strings"
	"wsocket"
)

var valid *validator.Validate

const jwtSecret = "Argusyes"

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	valid = validator.New()
}

func main() {
	defer mongoDB.Client.Close()

	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}

	ip := conf.Get("server.IP").(string)
	port := conf.Get("server.Port").(int64)
	allowOrigin := conf.Get("server.AllowOrigin").(string)
	addr := fmt.Sprintf("%s:%d", ip, port)

	router := gin.New()
	router.Use(ginAllowOriginMiddleware(allowOrigin))
	router.Use(ginAuthMiddleware())
	router.GET("/monitor", monitorHandler)
	router.GET("/ssh", sshHandler)
	router.POST("/user/addSSH", addUserSSHHandler)
	router.DELETE("/user/deleteSSH", deleteUserSSHHandler)
	router.PUT("/user/updateSSH", updateUserSSHHandler)
	router.GET("/user/selectSSH", selectUserSSHHandler)
	router.POST("/user/register", registerHandler)
	router.POST("/user/login", loginHandler)
	router.PUT("/user/changePasswd", changePasswdHandler)

	wsocket.WsocketManager.RegisterMessageHandler(messageRouter)
	wsocket.WsocketManager.RegisterCloseHandler(func(conn *wsocket.Connect) {
		ssh.Manager.ClearListener(conn.Key)
	})
	wsocket.WsocketManager.RegisterErrorHandler(func(conn *wsocket.Connect, err error) {
		ssh.Manager.ClearListener(conn.Key)
	})

	if err := router.Run(addr); err != nil {
		log.Fatal("failed run app: ", err)
	}
}

func monitorHandler(c *gin.Context) {
	wsocket.WsocketManager.HandleNewConnect(c.Writer, c.Request)
}

func sshHandler(c *gin.Context) {
	handleNewSSHConnect(c.Writer, c.Request)
}

func ginAllowOriginMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With, A-Token")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}

func ginAuthMiddleware() gin.HandlerFunc {

	abort := func(c *gin.Context, errText string) {
		c.AbortWithStatusJSON(http.StatusForbidden, Response{
			Code:    403,
			Message: &errText,
		})
	}

	return func(c *gin.Context) {
		if !isInWhiteList(c.Request.URL, c.Request.Method) {
			strToken := c.Request.Header.Get("A-Token")
			token, err := jwt.ParseWithClaims(strToken, &LoginRequest{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil {
				abort(c, fmt.Sprintf("Token auth fail : %v", err))
				return
			}
			loginRequest, ok := token.Claims.(*LoginRequest)
			if !ok {
				abort(c, fmt.Sprintf("Token auth fail : %v", err))
				return
			}
			if err := token.Claims.Valid(); err != nil {
				abort(c, fmt.Sprintf("Token auth fail : %v", err))
				return
			}
			c.Request.Header.Add("User-Name", loginRequest.UserName)
		}
		c.Next()
	}
}

func isInWhiteList(url *url.URL, method string) bool {
	whiteList := map[string]mapSet.Set[string]{
		"/user/register": mapSet.NewSet("POST"),
		"/user/login":    mapSet.NewSet("POST"),
		"/monitor":       mapSet.NewSet("GET"),
		"/ssh":           mapSet.NewSet("GET"),
	}
	queryUrl := strings.Split(fmt.Sprint(url), "?")[0]
	if set, ok := whiteList[queryUrl]; ok {
		if set.Contains(method) {
			return true
		}
		return false
	}
	return false
}
