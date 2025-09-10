package http

import (
	"fmt"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/gin-gonic/gin"
)

const (
	DefaultServiceTimeout = 3 * time.Second
)

type InitArgs struct {
	JWTSecret   []byte
	AuthService AuthService
}

func New(params InitArgs) (*gin.Engine, error) {
	if err := registerValidators(); err != nil {
		return nil, fmt.Errorf("initialize router: %w", err)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.Device())

	authHandler := NewAuthHandler(params.AuthService)

	api := r.Group("/api")
	api.POST("/login", middlewares.NonAuthRequired(params.JWTSecret), authHandler.Login)

	api.Use(middlewares.AuthRequired(params.JWTSecret))
	// JWT middleware protects all routes below of an api group
	return r, nil
}
