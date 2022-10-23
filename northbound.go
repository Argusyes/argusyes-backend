package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mongoDB"
	"net/http"
)

type Response struct {
	Code    string      `json:"code"`
	Message *string     `json:"message"`
	Data    interface{} `json:"data"`
}

type AddUserSSHRequest mongoDB.UserSSH

func addUserSSHHandler(context *gin.Context) {
	addUserSSHRequest := &AddUserSSHRequest{}
	err := context.BindJSON(addUserSSHRequest)
	if err != nil {
		errText := fmt.Sprintf("Request Parse Fail %v", err)
		context.JSON(http.StatusBadRequest, Response{
			Code:    "400",
			Message: &errText,
		})
	} else {
		if err := mongoDB.Client.InsertUserSSH(mongoDB.UserSSH(*addUserSSHRequest)); err == nil {
			context.JSON(http.StatusOK, Response{
				Code:    "200",
				Message: nil,
			})
		} else {
			failText := err.Error()
			context.JSON(http.StatusOK, Response{
				Code:    "400",
				Message: &failText,
			})
		}

	}
}
