package svcentry

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsdevblog/gophkeeper/internal/pag"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	"github.com/google/uuid"
)

const (
	defaultIsolationRetries = 3
	defaultIsolationTimeout = 30000
)

// Options for EntryService.
type Options struct {
	IsolationRetries uint // number of retries for transaction.
}
type EntryService struct {
	uow              uow.UOW
	isolationRetries uint
}

func New(uow uow.UOW, opts ...func(*Options)) *EntryService {
	options := &Options{
		IsolationRetries: defaultIsolationRetries,
	}
	for _, opt := range opts {
		opt(options)
	}
	return &EntryService{
		uow:              uow,
		isolationRetries: options.IsolationRetries,
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

func (e *EntryService) GetUserEntries(
	ctx context.Context,
	userID uuid.UUID,
	pagination *pag.Pagination,
) ([]models.Entry, int64, error) {
	var entries []models.Entry
	var totalEntries int64

	err := e.uow.DoWithIsolation(ctx, func(doCtx context.Context, tx uow.TX) error {
		entryRepo, errEntryRepo := uow.GetAs[EntryRepository](tx, uow.RepoName(repodto.EntryRepoName))
		if errEntryRepo != nil {
			return errEntryRepo //nolint:wrapcheck
		}

		fieldsRepo, errFieldsRepo := uow.GetAs[EntryFieldRepository](tx, uow.RepoName(repodto.EntryFieldRepoName))
		if errFieldsRepo != nil {
			return errFieldsRepo //nolint:wrapcheck
		}

		var errEntries error
		entries, errEntries = entryRepo.GetAllByUser(doCtx, userID, pagination.Limit(), pagination.Offset())

		if errEntries != nil {
			return errEntries //nolint:wrapcheck
		}

		var entryIDs = make([]uuid.UUID, len(entries))
		for i, entry := range entries {
			entryIDs[i] = entry.ID
		}

		fields, errFields := fieldsRepo.GetFieldsByEntryIDs(doCtx, entryIDs)
		if errFields != nil {
			return errFields //nolint:wrapcheck
		}

		var fieldsMap = make(map[uuid.UUID][]models.EntryField, len(entries))

		for _, field := range fields {
			fieldsMap[field.EntryID] = append(fieldsMap[field.EntryID], field)
		}
		for i, entry := range entries {
			if entryFields, ok := fieldsMap[entry.ID]; ok {
				entries[i].EntryFields = entryFields
			}
		}
		var errTotalRecords error
		totalEntries, errTotalRecords = entryRepo.GetCountByUser(doCtx, userID)
		if errTotalRecords != nil {
			return errTotalRecords // nolint:wrapcheck
		}

		return nil
	}, uow.IsolationDefault, e.isolationRetries)

	if err != nil {
		return nil, 0, fmt.Errorf("getting user entries: %w", err)
	}
	return entries, totalEntries, nil
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
		errBatch := fieldRepo.BatchCreate(
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

		if errBatch != nil {
			return errBatch //nolint:wrapcheck
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
