package sos_auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"net/http"
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

// Interacts with refresh token data source
var refreshNoMatch = errors.New("refresh token does not exist")

type CheckToken interface {
	Check(ctx context.Context, token string) (*RefreshData, error)
}

type CheckTokenRedis struct {
	Client *redis.Client
}

func (c *CheckTokenRedis) Check(ctx context.Context, token string) (*RefreshData, error) {
	res := c.Client.Get(token)
	if res.Err() != nil {
		if res.Err() == redis.Nil {
			return nil, refreshNoMatch
		}

		return nil, res.Err()
	}

	dataBytes, err := res.Bytes()
	if err != nil {
		return nil, err
	}

	data := &RefreshData{}
	if err := json.Unmarshal(dataBytes, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Handles the refresh token endpoint
type RefreshHandler struct {
	CheckToken CheckToken
	Secret     string
}

func (rh *RefreshHandler) Refresh(ctx *gin.Context) {
	// Get request body
	body := &RefreshBody{}
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.String(http.StatusBadRequest, "invalid format")
		return
	}

	// Check refresh token
	data, err := rh.CheckToken.Check(ctx, body.RefreshToken)
	if err != nil {
		if err == refreshNoMatch {
			ctx.String(http.StatusNotFound, "no match")
		} else {
			ctx.String(http.StatusInternalServerError, "internal error when checking refresh token")
		}

		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": data.UserId,
		"role":    data.Role,
	})

	tokenString, err := token.SignedString(rh.Secret)

	// Create Response
	ctx.JSON(http.StatusCreated, RefreshResponse{Token: tokenString})
}
