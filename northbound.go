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
	Code    int      `json:"code"`
	Message *string     `json:"message"`
	Data    interface{} `json:"data"`
}

type SelectUserSSHResponse struct {
	Code    int                      `json:"code"`
	Message *string                     `json:"message"`
	Data    []SelectUserSSHResponseData `json:"data"`
}

type SelectUserSSHResponseData struct {
	Name   string `json:"name"`
	Port   int    `json:"port"`
	Host   string `json:"host"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

type AddUserSSHResponse struct {
	Code    int                   `json:"code"`
	Message *string                  `json:"message"`
	Data    []AddUserSSHResponseData `json:"data"`
}

type AddUserSSHResponseData struct {
	Port  int    `json:"port"`
	Host  string `json:"host"`
	User  string `json:"user"`
	Added bool   `json:"added"`
}

type DeleteUserSSHResponse struct {
	Code    int                      `json:"code"`
	Message *string                     `json:"message"`
	Data    []DeleteUserSSHResponseData `json:"data"`
}

type DeleteUserSSHResponseData struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	User    string `json:"user"`
	Deleted bool   `json:"deleted"`
}

type UpdateUserSSHResponse struct {
	Code    int                      `json:"code"`
	Message *string                     `json:"message"`
	Data    []UpdateUserSSHResponseData `json:"data"`
}

type UpdateUserSSHResponseData struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	User    string `json:"user"`
	Updater bool   `json:"updater"`
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

type DeleteUserSSHRequest struct {
	Data []DeleteUserSSHRequestData `json:"data" validate:"required,dive"`
}

type DeleteUserSSHRequestData struct {
	Port int    `json:"port" validate:"required"`
	Host string `json:"host" validate:"required,ip_addr"`
	User string `json:"user" validate:"required"`
}

type UpdateUserSSHRequest struct {
	Data []UpdateUserSSHRequestData `json:"data" validate:"required,dive"`
}

type UpdateUserSSHRequestData struct {
	OldPort   int    `json:"oldPort" validate:"required"`
	OldHost   string `json:"oldHost" validate:"required,ip_addr"`
	OldUser   string `json:"oldUser" validate:"required"`
	NewPort   int    `json:"newPort" validate:"required"`
	NewHost   string `json:"newHost" validate:"required,ip_addr"`
	NewUser   string `json:"newUser" validate:"required"`
	NewName   string `json:"newName" validate:"required"`
	NewPasswd string `json:"newPasswd" validate:"required"`
}

type RegisterRequest struct {
	UserName string `json:"username" validate:"required"`
	Passwd   string `json:"passwd" validate:"required"`
}

type LoginRequest struct {
	jwt.StandardClaims
	UserName string `json:"username" validate:"required"`
	Passwd   string `json:"passwd" validate:"required"`
}

func requestJsonParseHelper(context *gin.Context, v interface{}) bool {
	if err := context.BindJSON(v); err != nil {
		errText := fmt.Sprintf("Request Json Parse Fail %v", err)
		context.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: &errText,
		})
		return false
	} else if err := valid.Struct(v); err != nil {
		errText := fmt.Sprintf("Request Validate Fail %v", err)
		context.JSON(http.StatusBadRequest, Response{
			Code:    400,
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
			Code:    200,
			Message: nil,
			Data:    make([]AddUserSSHResponseData, 0),
		}
		if err != nil {
			errText := fmt.Sprintf("Insert SSH Fail : %v", err)
			addUserSSHResponse.Code = 500
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

func deleteUserSSHHandler(context *gin.Context) {
	deleteUserSSHRequest := &DeleteUserSSHRequest{}
	if ok := requestJsonParseHelper(context, deleteUserSSHRequest); ok {
		username := context.Request.Header.Get("User-Name")
		userSSH := make([]mongoDB.UserSSH, 0)
		for _, ssh := range deleteUserSSHRequest.Data {
			userSSH = append(userSSH, mongoDB.UserSSH{
				UserName: username,
				Port:     ssh.Port,
				Host:     ssh.Host,
				User:     ssh.User,
			})
		}
		res, err := mongoDB.Client.DeleteUserSSH(userSSH)
		deleteUserSSHResponse := &DeleteUserSSHResponse{
			Code:    200,
			Message: nil,
			Data:    make([]DeleteUserSSHResponseData, 0),
		}
		if err != nil {
			errText := fmt.Sprintf("Delete SSH Fail : %v", err)
			deleteUserSSHResponse.Code = 500
			deleteUserSSHResponse.Message = &errText
		}
		resSet := mapSet.NewSet(res...)
		for _, ssh := range userSSH {
			if resSet.Contains(mongoDB.GeneralSSHId(ssh)) {
				deleteUserSSHResponse.Data = append(deleteUserSSHResponse.Data, DeleteUserSSHResponseData{
					Port:    ssh.Port,
					Host:    ssh.Host,
					User:    ssh.User,
					Deleted: true,
				})
			} else {
				deleteUserSSHResponse.Data = append(deleteUserSSHResponse.Data, DeleteUserSSHResponseData{
					Port:    ssh.Port,
					Host:    ssh.Host,
					User:    ssh.User,
					Deleted: false,
				})
			}
		}
		context.JSON(http.StatusOK, deleteUserSSHResponse)
	}
}

func updateUserSSHHandler(context *gin.Context) {
	updateUserSSHRequest := &UpdateUserSSHRequest{}
	if ok := requestJsonParseHelper(context, updateUserSSHRequest); ok {
		username := context.Request.Header.Get("User-Name")
		userSSHUpdater := make([]mongoDB.UserSSHUpdater, 0)
		for _, ssh := range updateUserSSHRequest.Data {
			userSSHUpdater = append(userSSHUpdater, mongoDB.UserSSHUpdater{
				OldSSH: mongoDB.UserSSH{
					UserName: username,
					Port:     ssh.OldPort,
					Host:     ssh.OldHost,
					User:     ssh.OldUser,
				},
				NewSSH: mongoDB.UserSSH{
					UserName: username,
					Port:     ssh.NewPort,
					Host:     ssh.NewHost,
					User:     ssh.NewUser,
					Passwd:   ssh.NewPasswd,
					Name:     ssh.NewName,
				},
			})
		}
		res, err := mongoDB.Client.UpdateUserSSH(userSSHUpdater)
		updateUserSSHResponse := &UpdateUserSSHResponse{
			Code:    200,
			Message: nil,
			Data:    make([]UpdateUserSSHResponseData, 0),
		}
		if err != nil {
			errText := fmt.Sprintf("Update SSH Fail : %v", err)
			updateUserSSHResponse.Code = 500
			updateUserSSHResponse.Message = &errText
		}
		resSet := mapSet.NewSet(res...)
		for _, u := range userSSHUpdater {
			if resSet.Contains(mongoDB.GeneralSSHId(u.OldSSH)) {
				updateUserSSHResponse.Data = append(updateUserSSHResponse.Data, UpdateUserSSHResponseData{
					Port:    u.OldSSH.Port,
					Host:    u.OldSSH.Host,
					User:    u.OldSSH.User,
					Updater: true,
				})
			} else {
				updateUserSSHResponse.Data = append(updateUserSSHResponse.Data, UpdateUserSSHResponseData{
					Port:    u.OldSSH.Port,
					Host:    u.OldSSH.Host,
					User:    u.OldSSH.User,
					Updater: false,
				})
			}
		}
		context.JSON(http.StatusOK, updateUserSSHResponse)
	}
}

func selectUserSSHHandler(context *gin.Context) {
	username := context.Request.Header.Get("User-Name")
	res, err := mongoDB.Client.SelectUserSSH(username)
	selectUserSSHResponse := &SelectUserSSHResponse{
		Code:    200,
		Message: nil,
		Data:    make([]SelectUserSSHResponseData, 0),
	}
	if err != nil {
		errText := fmt.Sprintf("Select SSH Fail : %v", err)
		selectUserSSHResponse.Code = 500
		selectUserSSHResponse.Message = &errText
	}
	for _, ssh := range res {
		selectUserSSHResponse.Data = append(selectUserSSHResponse.Data, SelectUserSSHResponseData{
			Port:   ssh.Port,
			Host:   ssh.Host,
			User:   ssh.User,
			Name:   ssh.Name,
			Passwd: ssh.Passwd,
		})
	}
	context.JSON(http.StatusOK, selectUserSSHResponse)
}

func registerHandler(context *gin.Context) {
	registerRequest := &RegisterRequest{}
	if ok := requestJsonParseHelper(context, registerRequest); ok {
		user := mongoDB.User{UserName: registerRequest.UserName, Passwd: registerRequest.Passwd}
		if err := mongoDB.Client.InsertUser(user); err != nil {
			errText := fmt.Sprintf("Insert User Fail %v", err)
			context.JSON(http.StatusOK, Response{
				Code:    500,
				Message: &errText,
			})
		} else {
			context.JSON(http.StatusOK, Response{
				Code:    200,
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
				Code:    500,
				Message: &errText,
			})
		} else {
			loginRequest.IssuedAt = time.Now().Unix()
			loginRequest.ExpiresAt = time.Now().Add(time.Second * time.Duration(3600*24*3)).Unix()
			loginRequest.Passwd = "Argusyes"
			if token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, loginRequest).SignedString([]byte(jwtSecret)); err != nil {
				errText := fmt.Sprintf("Create User Token Fail %v", err)
				context.JSON(http.StatusOK, Response{
					Code:    500,
					Message: &errText,
				})
			} else if err := mongoDB.Client.UpdateUserToken(user.UserName, token); err != nil {
				errText := fmt.Sprintf("Update User Token Fail %v", err)
				context.JSON(http.StatusOK, Response{
					Code:    500,
					Message: &errText,
				})
			} else {
				context.JSON(http.StatusOK, Response{
					Code:    200,
					Message: nil,
					Data:    token,
				})
			}
		}
	}
}
