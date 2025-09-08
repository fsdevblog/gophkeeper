package svcentry

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	"github.com/google/uuid"
)

type EntryService struct {
	uow uow.UOW
}

func New(uow uow.UOW) *EntryService {
	return &EntryService{
		uow: uow,
	}
}

type EntryFieldArg struct {
	Key       models.EntryFieldKeyType
	Value     []byte
	IsPrivate bool
}

type CreateEntryArgs struct {
	Title       string
	UserID      uuid.UUID
	DeviceID    uuid.UUID
	EntryType   models.EntryType
	EntryFields []EntryFieldArg
}

func (e *EntryService) CreateEntry(ctx context.Context, args CreateEntryArgs) (*models.Entry, error) {
	var entry *models.Entry
	err := e.uow.Do(ctx, func(doCtx context.Context, tx uow.TX) error {
		entryRepo, errEntryRepo := uow.GetAs[EntryRepository](tx, uow.RepoName(repodto.EntryRepoName))
		if errEntryRepo != nil {
			return errEntryRepo //nolint:wrapcheck
		}
		var errCreate error
		entry, errCreate = entryRepo.Create(doCtx, repodto.CreateEntryArgs{
			Title:     args.Title,
			UserID:    args.UserID,
			DeviceID:  args.DeviceID,
			EntryType: args.EntryType,
		})
		if errCreate != nil {
			return errCreate //nolint:wrapcheck
		}

		fieldRepo, errFieldRepo := uow.GetAs[EntryFieldRepository](tx, uow.RepoName(repodto.EntryFieldRepoName))
		if errFieldRepo != nil {
			return errFieldRepo //nolint:wrapcheck
		}

		var errIntoBatch error
		var fields2Create = make([]repodto.CreateEntryFieldArgs, len(args.EntryFields))
		for i, field := range args.EntryFields {
			fields2Create[i] = repodto.CreateEntryFieldArgs{
				EntryID: entry.ID,
				Key:     field.Key,
				Value:   field.Value,
			}
		}
		var createdFields = make([]models.EntryField, len(args.EntryFields))
		errFieldBatch := fieldRepo.BatchCreate(
			doCtx,
			entry.ID,
			fields2Create,
			func(i int, field *models.EntryField, err error) {
				if err != nil {
					errIntoBatch = errors.Join(errIntoBatch, err)
					return
				}
				createdFields[i] = *field
			},
		)

		if errFieldBatch != nil {
			return errFieldBatch //nolint:wrapcheck
		}
		if errIntoBatch != nil {
			return errIntoBatch //nolint:wrapcheck
		}

		entry.EntryFields = createdFields
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("creating entry: %w", err)
	}
	return entry, nil
}
