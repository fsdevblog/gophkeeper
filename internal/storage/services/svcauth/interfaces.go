package svcauth

import (
	"context"

	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

// UserRepository интерфейс юзер репозитория.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, args repodto.CreateUserArgs) (*models.User, error)
}

// PasswordHasher интерфейс хешера пароля.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(password string, hashedPassword string) bool
}
