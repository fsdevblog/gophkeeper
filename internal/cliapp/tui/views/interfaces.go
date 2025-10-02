package views

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/models"
)

type AuthProvider interface {
	Login(ctx context.Context, username string, password string) (*models.User, error)
	Register(ctx context.Context, username string, password string) (*models.User, error)
}
