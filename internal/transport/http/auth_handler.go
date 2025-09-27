package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/dto"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService AuthProvider
}

func NewAuthHandler(authService AuthProvider) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (a *AuthHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (a *AuthHandler) Login(c *gin.Context) {
	params, bindOk := bindAs[dto.AuthenticateParams](c)
	if !bindOk {
		return
	}
	deviceHash, _ := c.Get(middlewares.DeviceHashContextKey)

	ctx, cancel := context.WithTimeout(c, DefaultServiceTimeout)
	defer cancel()

	token, user, err := a.authService.Authenticate(ctx, svcauth.AuthenticateArgs{
		Username: strings.TrimSpace(params.Username),
		Password: strings.TrimSpace(params.Password),
		Device: svcauth.DeviceArgs{
			DeviceType: params.Device.DeviceType,
			// no need to check return value bcoz middleware checks it.
			DeviceHash:      deviceHash.(uuid.UUID), //nolint:errcheck
			Platform:        params.Device.Platform,
			PlatformVersion: params.Device.PlatformVersion,
			AppVersion:      params.Device.AppVersion,
		},
	})
	if err != nil {
		if errors.Is(err, svcauth.ErrInvalidCredentials) {
			_ = c.AbortWithError(http.StatusUnauthorized, errors.New("invalid credentials")).
				SetType(gin.ErrorTypePublic)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err).
			SetType(gin.ErrorTypePrivate)
		return
	}
	c.Header("Authorization", "Bearer "+token)

	c.JSON(http.StatusOK, gin.H{"user": dto.UserItem{
		ID:       user.ID,
		Username: user.Username,
	}})
}

func (a *AuthHandler) Register(c *gin.Context) {
	params, bindOk := bindAs[dto.RegisterParams](c)
	if !bindOk {
		return
	}

	deviceHash, _ := c.Get(middlewares.DeviceHashContextKey)

	ctx, cancel := context.WithTimeout(c, DefaultServiceTimeout)
	defer cancel()
	token, user, err := a.authService.Register(ctx, svcauth.RegisterArgs{
		Username: strings.TrimSpace(params.Username),
		Password: strings.TrimSpace(params.Password),
		Device: svcauth.DeviceArgs{
			DeviceType:      params.Device.DeviceType,
			DeviceHash:      deviceHash.(uuid.UUID), //nolint:errcheck
			Platform:        params.Device.Platform,
			PlatformVersion: params.Device.PlatformVersion,
			AppVersion:      params.Device.AppVersion,
		},
	})
	if err != nil {
		if errors.Is(err, svcauth.ErrUserAlreadyRegistered) {
			_ = c.AbortWithError(http.StatusConflict, errors.New("user already registered")).
				SetType(gin.ErrorTypePublic)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err).
			SetType(gin.ErrorTypePrivate)
		return
	}
	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"user": dto.UserItem{
		ID:       user.ID,
		Username: user.Username,
	}})
}

func bindAs[T any](c *gin.Context) (*T, bool) {
	var params T
	if errBind := c.ShouldBindJSON(&params); errBind != nil {
		var errValidator validator.ValidationErrors
		if errors.As(errBind, &errValidator) {
			_ = c.Error(errBind).
				SetType(gin.ErrorTypeBind)
			c.Status(http.StatusUnprocessableEntity)
			return nil, false
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, false
	}
	return &params, true
}
