package svcentry

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

type EntryRepository interface {
	Create(ctx context.Context, args repodto.CreateEntryArgs) (*models.Entry, error)
	GetAllByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]models.Entry, error)
	GetCountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type EntryFieldRepository interface {
	BatchCreate(
		ctx context.Context,
		entryID uuid.UUID,
		fields []repodto.CreateEntryFieldArgs,
		resultRow func(i int, field *models.EntryField, err error),
	) error
	GetFieldsByEntryIDs(ctx context.Context, entryIDs []uuid.UUID) ([]models.EntryField, error)
}
