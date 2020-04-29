package sos_auth

import (
	"context"
	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CheckPassword interface {
	Check(ctx context.Context, username string, password string) error
}

type AuthHandler struct {
	CheckPassword CheckPassword
}

func (ah *AuthHandler) Authenticate(ctx *gin.Context) {
	panic("implement me")
}
