package main

import (
	"fmt"
	mapSet "github.com/deckarep/golang-set/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"mongoDB"
	"net/http"
	"time"
)

type Response struct {
	Code    string      `json:"code"`
	Message *string     `json:"message"`
	Data    interface{} `json:"data"`
}

type AddUserSSHResponse struct {
	Code    string                   `json:"code"`
	Message *string                  `json:"message"`
	Data    []AddUserSSHResponseData `json:"data"`
}

type AddUserSSHResponseData struct {
	Port  int    `json:"port"`
	Host  string `json:"host"`
	User  string `json:"user"`
	Added bool   `json:"added"`
}

type AddUserSSHRequest struct {
	Data []AddUserSSHRequestData `json:"data" validate:"required,dive"`
}

type AddUserSSHRequestData struct {
	Name   string `json:"name" validate:"required"`
	Port   int    `json:"port" validate:"required"`
	Host   string `json:"host" validate:"required,ip_addr"`
	User   string `json:"user" validate:"required"`
	Passwd string `json:"passwd" validate:"required"`
}

type RegisterRequest struct {
	UserName string `json:"user_name" validate:"required"`
	Passwd   string `json:"passwd" validate:"required"`
}

type LoginRequest struct {
	jwt.StandardClaims
	UserName string `json:"user_name" validate:"required"`
	Passwd   string `json:"passwd" validate:"required"`
}

func requestJsonParseHelper(context *gin.Context, v interface{}) bool {
	if err := context.BindJSON(v); err != nil {
		errText := fmt.Sprintf("Request Json Parse Fail %v", err)
		context.JSON(http.StatusBadRequest, Response{
			Code:    "400",
			Message: &errText,
		})
		return false
	} else if err := valid.Struct(v); err != nil {
		errText := fmt.Sprintf("Request Validate Fail %v", err)
		context.JSON(http.StatusBadRequest, Response{
			Code:    "400",
			Message: &errText,
		})
		return false
	}
	return true
}

func addUserSSHHandler(context *gin.Context) {
	addUserSSHRequest := &AddUserSSHRequest{}
	if ok := requestJsonParseHelper(context, addUserSSHRequest); ok {
		username := context.Request.Header.Get("User-Name")
		userSSH := make([]mongoDB.UserSSH, 0)
		for _, ssh := range addUserSSHRequest.Data {
			userSSH = append(userSSH, mongoDB.UserSSH{
				UserName: username,
				Name:     ssh.Name,
				Port:     ssh.Port,
				Host:     ssh.Host,
				User:     ssh.User,
				Passwd:   ssh.Passwd,
			})
		}
		res, err := mongoDB.Client.InsertUserSSH(userSSH)
		addUserSSHResponse := &AddUserSSHResponse{
			Code:    "200",
			Message: nil,
			Data:    make([]AddUserSSHResponseData, 0),
		}
		if err != nil {
			errText := fmt.Sprintf("Insert SSH Fail : %v", err)
			addUserSSHResponse.Code = "500"
			addUserSSHResponse.Message = &errText
		}
		resSet := mapSet.NewSet(res...)
		for _, ssh := range userSSH {
			if resSet.Contains(mongoDB.GeneralSSHId(ssh)) {
				addUserSSHResponse.Data = append(addUserSSHResponse.Data, AddUserSSHResponseData{
					Port:  ssh.Port,
					Host:  ssh.Host,
					User:  ssh.User,
					Added: true,
				})
			} else {
				addUserSSHResponse.Data = append(addUserSSHResponse.Data, AddUserSSHResponseData{
					Port:  ssh.Port,
					Host:  ssh.Host,
					User:  ssh.User,
					Added: false,
				})
			}
		}
		context.JSON(http.StatusOK, addUserSSHResponse)
	}
}

func registerHandler(context *gin.Context) {
	registerRequest := &RegisterRequest{}
	if ok := requestJsonParseHelper(context, registerRequest); ok {
		user := mongoDB.User{UserName: registerRequest.UserName, Passwd: registerRequest.Passwd}
		if err := mongoDB.Client.InsertUser(user); err != nil {
			errText := fmt.Sprintf("Insert User Fail %v", err)
			context.JSON(http.StatusOK, Response{
				Code:    "500",
				Message: &errText,
			})
		} else {
			context.JSON(http.StatusOK, Response{
				Code:    "200",
				Message: nil,
			})
		}
	}
}

func loginHandler(context *gin.Context) {
	loginRequest := &LoginRequest{}
	if ok := requestJsonParseHelper(context, loginRequest); ok {
		user := mongoDB.User{UserName: loginRequest.UserName, Passwd: loginRequest.Passwd}
		if err := mongoDB.Client.CheckUserPasswd(user); err != nil {
			errText := fmt.Sprintf("Check User Fail %v", err)
			context.JSON(http.StatusOK, Response{
				Code:    "500",
				Message: &errText,
			})
		} else {
			loginRequest.IssuedAt = time.Now().Unix()
			loginRequest.ExpiresAt = time.Now().Add(time.Second * time.Duration(3600*24*3)).Unix()
			loginRequest.Passwd = "Argusyes"
			if token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, loginRequest).SignedString([]byte(jwtSecret)); err != nil {
				errText := fmt.Sprintf("Create User Token Fail %v", err)
				context.JSON(http.StatusOK, Response{
					Code:    "500",
					Message: &errText,
				})
			} else if err := mongoDB.Client.UpdateUserToken(user.UserName, token); err != nil {
				errText := fmt.Sprintf("Update User Token Fail %v", err)
				context.JSON(http.StatusOK, Response{
					Code:    "500",
					Message: &errText,
				})
			} else {
				context.JSON(http.StatusOK, Response{
					Code:    "200",
					Message: nil,
					Data:    token,
				})
			}
		}
	}
}
