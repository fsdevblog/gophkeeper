package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// DeviceParams fields for device registration. DeviceHash must present in each HTTP header.
type DeviceParams struct {
	DeviceType      models.DeviceType `binding:"required" json:"deviceType"`
	Platform        string            `binding:"required,max=32" json:"platform"`
	PlatformVersion string            `binding:"required,max=16" json:"platformVersion"`
	AppVersion      string            `binding:"required,max=16" json:"appVersion"`
}

type AuthenticateParams struct {
	Username string       `binding:"required,min=1,max=15" json:"username"`
	Password string       `binding:"required,min=6,max_bytes=72" json:"password"`
	Device   DeviceParams `binding:"required"              json:"device"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

func (a *AuthHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (a *AuthHandler) Login(c *gin.Context) {
	var params AuthenticateParams
	deviceHash, _ := c.Get(middlewares.DeviceHashContextKey)

	if errBind := c.ShouldBindJSON(&params); errBind != nil {
		var errValidator validator.ValidationErrors
		if errors.As(errBind, &errValidator) {
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errBind).
				SetType(gin.ErrorTypeBind)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"user": UserResponse{
		ID:       user.ID,
		Username: user.Username,
	}})
}

type RegisterParams struct {
	Username string       `binding:"required,min=1,max=15" json:"username"`
	Password string       `binding:"required,min=6,max_bytes=72" json:"password"`
	Device   DeviceParams `binding:"required"              json:"device"`
}

func (a *AuthHandler) Register(c *gin.Context) {
	var params RegisterParams
	deviceHash, _ := c.Get(middlewares.DeviceHashContextKey)
	if errBind := c.ShouldBindJSON(&params); errBind != nil {
		var errValidator validator.ValidationErrors
		if errors.As(errBind, &errValidator) {
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errBind).
				SetType(gin.ErrorTypeBind)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(c, DefaultServiceTimeout)
	defer cancel()
	token, user, err := a.authService.Register(ctx, svcauth.RegisterArgs{
		Username: strings.TrimSpace(params.Username),
		Password: strings.TrimSpace(params.Password),
		Device: svcauth.DeviceArgs{
			DeviceType:      params.Device.DeviceType,
			DeviceHash:      deviceHash.(uuid.UUID),
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
	c.JSON(http.StatusOK, gin.H{"user": UserResponse{
		ID:       user.ID,
		Username: user.Username,
	}})
}
