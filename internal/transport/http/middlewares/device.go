package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	DeviceHashHeaderKey  = "X-Device-Hash"
	DeviceHashContextKey = "deviceHash"
)

// Device middleware catches the device hash from the request header.
// If the hash is invalid, the middleware aborts the request.
func Device() gin.HandlerFunc {
	return func(c *gin.Context) {
		hashStr := c.GetHeader(DeviceHashHeaderKey)
		hashUUID, errParse := uuid.Parse(hashStr)
		if errParse != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("invalid device hash")).
				SetType(gin.ErrorTypePublic)
			return
		}
		c.Set(DeviceHashContextKey, hashUUID)
		c.Next()
	}
}
