package sos_auth

import (
	"context"
	"github.com/gin-gonic/gin"
)

// Request body for refresh endpoint
type RefreshBody struct {
	RefreshToken string `json:"refresh_token" binding:"required,uuid4"`
}

// Refresh response for refresh endpoint
type RefreshResponse struct {
	Token string `json:"token"`
}

// Actual data store within data storage
type RefreshData struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
}

type CheckToken interface {
	Check(ctx context.Context, token string) (RefreshData, error)
}

type RefreshHandler struct {
	CheckToken CheckToken
	Secret     string
}

func (rh *RefreshHandler) Refresh(ctx *gin.Context) {
	panic("implement me")
}
