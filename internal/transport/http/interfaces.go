package http

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/pag"
	"github.com/google/uuid"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

type AuthProvider interface {
	Authenticate(ctx context.Context, args svcauth.AuthenticateArgs) (string, *models.User, error)
	Register(ctx context.Context, args svcauth.RegisterArgs) (string, *models.User, error)
}

type EntryProvider interface {
	GetUserEntries(ctx context.Context, userID uuid.UUID, p *pag.Pagination) ([]models.Entry, int64, error)
	// GetSafeEntryFields Must hide private fields.
	GetSafeEntryFields(ctx context.Context, userID uuid.UUID, entryID uuid.UUID) ([]models.EntryField, error)
}
