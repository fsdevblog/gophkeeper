package http

import (
	"context"
	"errors"
	"net/http"

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

// DeviceArgs fields for device registration. DeviceHash must present in each HTTP header.
type DeviceArgs struct {
	DeviceType      models.DeviceType `json:"deviceType"`
	Platform        string            `json:"platform"`
	PlatformVersion string            `json:"platformVersion"`
	AppVersion      string            `json:"appVersion"`
}

type AuthenticateArgs struct {
	Username string     `binding:"required,min=1,max=15" json:"username"`
	Password string     `binding:"required,min=6,max=72" json:"password"`
	Device   DeviceArgs `binding:"required"              json:"device"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

func (a *AuthHandler) Login(c *gin.Context) {
	var params AuthenticateArgs
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
		Username: params.Username,
		Password: params.Password,
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
