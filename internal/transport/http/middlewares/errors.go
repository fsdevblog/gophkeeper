package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func statusErrorText(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not found"
	case http.StatusUnprocessableEntity:
		return "unprocessable entity"
	case http.StatusConflict:
		return "conflict"
	default:
		return "internal server error"
	}
}

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		firstErr := c.Errors[0]
		var msg string
		//nolint:exhaustive
		switch firstErr.Type {
		case gin.ErrorTypePublic, gin.ErrorTypeBind:
			msg = firstErr.Error()
		default:
			msg = statusErrorText(c.Writer.Status())
		}

		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

		c.JSON(c.Writer.Status(), gin.H{"error": msg})
		c.Abort()
	}
}
