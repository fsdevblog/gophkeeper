package http

import (
	"fmt"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/storage/services"
	"go.uber.org/zap"

	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/gin-gonic/gin"
)

const (
	DefaultServiceTimeout = 3 * time.Second
)

type InitArgs struct {
	JWTSecret []byte
	Services  *services.Collection
	Logger    *zap.Logger
}

func New(params InitArgs) (*gin.Engine, error) {
	if err := registerValidators(); err != nil {
		return nil, fmt.Errorf("initialize router: %w", err)
	}

	authHandler := NewAuthHandler(params.Services.AuthService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.LoggerMiddleware(params.Logger))
	r.GET("/ping", authHandler.Ping)

	r.Use(middlewares.Device())

	api := r.Group("/api")
	api.POST("/login", middlewares.NonAuthRequired(params.JWTSecret), authHandler.Login)

	api.Use(middlewares.AuthRequired(params.JWTSecret))
	// JWT middleware protects all routes below of an api group
	return r, nil
}
