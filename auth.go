package sos_auth

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Request body
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Interacts with data source
var userNoMatch = errors.New("username and password does not match")

type CheckPassword interface {
	Check(ctx context.Context, username string, password string) error
}

type CheckPasswordPq struct {
	Pool *pgxpool.Pool
}

func (cp *CheckPasswordPq) Check(ctx context.Context, username string, password string) error {
	// Get connection from Pool
	conn, err := cp.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Execute query
	rows, err := conn.Query(ctx, "SELECT 1 FROM User WHERE username = $1 AND password = $2 LIMIT 1", username, password)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Check that record is found
	count := 0
	for rows.Next() {
		count += 1
	}

	if count != 1 {
		return userNoMatch
	}

	return nil
}

// Gin handler
type AuthHandler struct {
	CheckPassword CheckPassword
}

func (ah *AuthHandler) Authenticate(ctx *gin.Context) {
	panic("implement me")
}
