package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fsdevblog/gophkeeper/internal/tokens"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CurrentUserIDKey is the context key for storing the current user's ID.
const CurrentUserIDKey = "currentUserID"

// checkAuthorization extracts and validates the JWT token from the Authorization header.
//
// Parameters:
//   - c: Gin context
//   - jwtTokenSecret: secret key for JWT token validation
//
// Returns:
//   - *jwt.Token: validated token on successful validation
//   - error: error if token is missing or invalid. If a token does not exist, ErrTokenNotExist is returned.
//
// The token must be provided in the header in "Bearer <token>" format.
func checkAuthorization(c *gin.Context, jwtTokenSecret []byte) (*jwt.Token, error) {
	tokenHeader := c.GetHeader("Authorization")
	bearer := "Bearer "

	if len(tokenHeader) < len(bearer) || tokenHeader[:len(bearer)] != bearer {
		return nil, ErrTokenNotExist
	}

	tokenStr := tokenHeader[len(bearer):]
	token, err := tokens.ValidateUserJWT(tokenStr, jwtTokenSecret)
	if err != nil {
		return nil, fmt.Errorf("check authorization: %w", err)
	}
	return token, nil
}

// AuthRequired returns a Gin middleware function for request authorization.
// It validates the JWT token from the Authorization header and stores the user ID
// in the request context under CurrentUserIDKey.
//
// Parameters:
//   - jwtTokenSecret: secret key for JWT token validation
//
// The request is aborted with 401 Unauthorized if the token is missing or invalid.
// If the claims format is incorrect, the request is aborted with 500 Internal Server Error.
func AuthRequired(jwtTokenSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := checkAuthorization(c, jwtTokenSecret)
		if err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, errors.New("auth required")).
				SetType(gin.ErrorTypePublic)
			if !errors.Is(err, ErrTokenNotExist) {
				_ = c.Error(err).SetType(gin.ErrorTypePrivate)
			}
			return
		}
		userClaim, ok := token.Claims.(*tokens.UserClaims)
		if !ok {
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("invalid jwt claims type")).
				SetType(gin.ErrorTypePrivate)
			return
		}
		c.Set(CurrentUserIDKey, userClaim.ID)
		c.Next()
	}
}

// NonAuthRequired returns a Gin middleware function for handling requests
// that require the absence of authorization.
//
// Parameters:
//   - jwtTokenSecret: secret key for JWT token validation
//
// If the request contains a valid authorization token, it is aborted
// with 401 Unauthorized and the message "you are already logged in".
func NonAuthRequired(jwtTokenSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := checkAuthorization(c, jwtTokenSecret)
		if err == nil {
			_ = c.AbortWithError(http.StatusUnauthorized, errors.New("you are already logged in")).
				SetType(gin.ErrorTypePublic)
			return
		}

		c.Next()
	}
}
