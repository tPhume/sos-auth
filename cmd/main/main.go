package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tPhume/sos-auth"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	var redisAddr, redisUri string
	// Get env
	secret := os.Getenv("JWT_SECRET")
	psql := os.Getenv("PSQL_URI")

	redisAddr = os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisUri = os.Getenv("REDIS_URI")
	}

	failOnEmpty(secret, psql, redisUri)

	byteSecret := []byte(secret)

	// Get Postgres Connection Pool
	pool, err := pgxpool.Connect(context.Background(), psql)
	failOnError("could not get pgx connection pool", err)

	// Get Redis client
	var redisOpt *redis.Options

	if redisUri != "" {
		redisOpt, err = redis.ParseURL(redisUri)
		failOnError("bad redis uri", err)
	} else {
		redisOpt = &redis.Options{Addr: redisAddr, DB: 0}
	}

	redisClient := redis.NewClient(redisOpt)

	// Create Repo
	checkPassword := &auth.CheckPasswordPq{Pool: pool}
	addRefreshToken := &auth.AddRefreshTokenRedis{Client: redisClient}
	checkToken := &auth.CheckTokenRedis{Client: redisClient}

	// Create server
	corsConfig := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	engine := gin.New()
	engine.Use(cors.New(corsConfig))
	authHandler := &auth.AuthHandler{CheckPassword: checkPassword, AddRefreshToken: addRefreshToken, Secret: byteSecret}
	engine.POST("/api/v1/authenticate", authHandler.Authenticate)

	refreshHandler := &auth.RefreshHandler{CheckToken: checkToken, Secret: byteSecret}
	engine.POST("/api/v1/refresh", refreshHandler.Refresh)

	log.Fatal(engine.Run("0.0.0.0:4356"))
}

func failOnEmpty(env ...string) {
	for _, v := range env {
		if strings.TrimSpace(v) == "" {
			log.Fatal("missing env")
		}
	}
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
