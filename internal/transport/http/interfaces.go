package http

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

type AuthService interface {
	Authenticate(ctx context.Context, args svcauth.AuthenticateArgs) (string, *models.User, error)
	Register(ctx context.Context, args svcauth.RegisterArgs) (string, *models.User, error)
}
