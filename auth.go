package auth

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
	"golang.org/x/crypto/bcrypt"
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
var UserNoMatch = errors.New("email and password does not match")

type CheckPassword interface {
	Check(ctx context.Context, user *User) error
}

type CheckPasswordPq struct {
	Pool *pgxpool.Pool
}

func (cp *CheckPasswordPq) Check(ctx context.Context, user *User) error {
	// Get connection from Pool
	conn, err := cp.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Execute query
	temp := user.Password

	query := "SELECT id, role, username, password FROM \"User\" WHERE email = $1 LIMIT 1"
	if err := conn.QueryRow(ctx, query, user.Email).Scan(&user.UserId, &user.Role, &user.Name, &user.Password); err != nil {
		if err == pgx.ErrNoRows {
			return UserNoMatch
		}

		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(temp)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return UserNoMatch
		}

		return err
	}

	return nil
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
	Secret          []byte
}

func (ah *AuthHandler) Authenticate(ctx *gin.Context) {
	// Get request body
	user := &User{}
	if err := ctx.ShouldBindJSON(user); err != nil {
		ctx.String(http.StatusBadRequest, "invalid format")
		return
	}

	// Check is username and password match
	if err := ah.CheckPassword.Check(ctx, user); err != nil {
		if err == UserNoMatch {
			ctx.String(http.StatusNotFound, "no match")
		} else {
			ctx.String(http.StatusInternalServerError, "internal error when checking password, error: %s", err.Error())
		}

		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserId,
		"role":    user.Role,
	})

	tokenString, err := token.SignedString(ah.Secret)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "internal error when creating jwt token: %s", err.Error())
		return
	}

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
