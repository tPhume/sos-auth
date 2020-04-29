package sos_auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"time"
)

// User entity
type User struct {
	UserId   int
	Role     string
	Name     string
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Data to store with refresh token
type RefreshData struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
}

// Response body
type AuthResponse struct {
	UserId       int    `json:"user_id"`
	Role         string `json:"role"`
	Name         string `json:"name"`
	Email        string `json:"email" binding:"required"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Interacts with user data source
var userNoMatch = errors.New("email and password does not match")

type CheckPassword interface {
	Check(ctx context.Context, username string, password string) (*User, error)
}

type CheckPasswordPq struct {
	Pool *pgxpool.Pool
}

func (cp *CheckPasswordPq) Check(ctx context.Context, email string, password string) (*User, error) {
	// Get connection from Pool
	conn, err := cp.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Execute query
	user := &User{Email: email, Password: password}

	query := "SELECT user_id, role, name FROM User WHERE email = $1 AND password = $2 LIMIT 1"
	if err := conn.QueryRow(ctx, query, email, password).Scan(&user.UserId, &user.Role, &user.Name); err != nil {
		if err == pgx.ErrNoRows {
			return nil, userNoMatch
		}

		return nil, err
	}

	return user, nil
}

// Interacts with refresh token data source
type AddRefreshToken interface {
	Add(ctx context.Context, refreshToken string, data RefreshData) error
}

type AddRefreshTokenRedis struct {
	Client *redis.Client
}

func (a *AddRefreshTokenRedis) Add(ctx context.Context, refreshToken string, data RefreshData) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if status := a.Client.Set(refreshToken, string(dataJSON), time.Hour*8); status.Err() != nil {
		return status.Err()
	}

	return nil
}

// Gin handler
type AuthHandler struct {
	CheckPassword   CheckPassword
	AddRefreshToken AddRefreshToken
	Secret          string
}

func (ah *AuthHandler) Authenticate(ctx *gin.Context) {
	// Get request body
	user := &User{}
	if err := ctx.ShouldBindJSON(user); err != nil {
		ctx.String(http.StatusBadRequest, "invalid format")
		return
	}

	// TODO Add hashing algorithm for password if necessary

	// Check is username and password match
	user, err := ah.CheckPassword.Check(ctx, user.Email, user.Password)
	if err != nil {
		if err == userNoMatch {
			ctx.String(http.StatusNotFound, "no match")
		} else {
			ctx.String(http.StatusInternalServerError, "internal error when checking password")
		}

		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserId,
		"role":    user.Role,
	})

	tokenString, err := token.SignedString(ah.Secret)

	// Add refresh token to Redis
	refreshToken := uuid.New().String()
	if err := ah.AddRefreshToken.Add(ctx, refreshToken, RefreshData{UserId: user.UserId, Role: user.Role}); err != nil {
		ctx.String(http.StatusInternalServerError, "internal error when creating refresh token")
		return
	}

	// Create Response
	response := &AuthResponse{
		UserId:       user.UserId,
		Role:         user.Role,
		Name:         user.Name,
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
	}

	ctx.JSON(http.StatusCreated, response)
}
