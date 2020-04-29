package sos_auth

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
)

// User entity
type User struct {
	UserId   string
	Role     string
	Name     string
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Response body
type Token struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Interacts with data source
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

// Gin handler
type AuthHandler struct {
	CheckPassword CheckPassword
	Secret        string
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
			ctx.String(http.StatusInternalServerError, "internal error")
		}

		return
	}

	// TODO Create JWT token

	// TODO Add new refresh token to Redis
}
