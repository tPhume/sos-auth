package sos_auth

import (
	"context"
	"github.com/gin-gonic/gin"
)

type Refresh struct {
	RefreshToken string `json:"refresh_token" binding:"required,uuid4"`
}

type CheckToken interface {
	Check(ctx context.Context, token string) error
}

type RefreshHandler struct {
	CheckToken CheckToken
}

func (rh *RefreshHandler) Refresh(ctx *gin.Context) {
	panic("implement me")
}
