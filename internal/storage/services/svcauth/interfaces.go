package svcauth

import (
	"context"

	"github.com/google/uuid"

	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

// UserRepository интерфейс юзер репозитория.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, args repodto.CreateUserArgs) (*models.User, error)
}

type DeviceRepository interface {
	Create(ctx context.Context, userID uuid.UUID, args repodto.CreateDeviceArgs) error
}

// PasswordHasher хешер пароля.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(password string, hashedPassword string) bool
}
