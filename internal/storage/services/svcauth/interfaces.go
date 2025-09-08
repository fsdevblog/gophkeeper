package svcauth

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

// UserRepository интерфейс юзер репозитория.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
}

// PasswordHasher интерфейс хешера пароля.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(password string, hashedPassword string) bool
}
