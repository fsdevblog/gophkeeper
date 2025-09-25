package http

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/gin-gonic/gin"
)

const (
	DefaultServiceTimeout = 3 * time.Second
)

type InitArgs struct {
	JWTSecret     []byte
	AuthProvider  AuthProvider
	EntryProvider EntryProvider
	Logger        *zap.Logger
}

func MustNew(params InitArgs) *gin.Engine {
	r, err := New(params)
	if err != nil {
		panic(err)
	}
	return r
}

func New(params InitArgs) (*gin.Engine, error) {
	if err := registerValidators(); err != nil {
		return nil, fmt.Errorf("initialize router: %w", err)
	}

	authHandler := NewAuthHandler(params.AuthProvider)
	entryHandler := NewEntriesHandler(params.EntryProvider)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.LoggerMiddleware(params.Logger))
	r.GET("/ping", authHandler.Ping)

	r.Use(middlewares.Device())

	api := r.Group("/api")
	api.POST("/auth/login", middlewares.NonAuthRequired(params.JWTSecret), authHandler.Login)
	api.POST("/auth/register", middlewares.NonAuthRequired(params.JWTSecret), authHandler.Register)

	api.Use(middlewares.AuthRequired(params.JWTSecret))
	// JWT middleware protects all routes below of an api group
	api.GET("/entries", entryHandler.GetAll)
	api.GET("/entries/:entryID/fields", entryHandler.GetEntryFields)
	return r, nil
}
