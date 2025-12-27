package pgrepo

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/sqlcgen"
	"github.com/google/uuid"
)

type EntryFieldRepo struct {
	q *sqlcgen.Queries
}

func NewEntryFieldRepo(conn sqlcgen.DBTX) *EntryFieldRepo {
	return &EntryFieldRepo{q: sqlcgen.New(conn)}
}

func (e *EntryFieldRepo) BatchCreate(
	ctx context.Context,
	fields []repodto.CreateEntryFieldArgs,
	resultRow func(i int, field *models.EntryField, err error),
) error {
	args := make([]sqlcgen.EntryFields_BatchCreateParams, len(fields))
	for i, field := range fields {
		args[i] = sqlcgen.EntryFields_BatchCreateParams{
			EntryID:   field.EntryID,
			Key:       field.Key,
			Value:     field.Value,
			IsPrivate: field.IsPrivate,
		}
	}
	results := e.q.EntryFields_BatchCreate(ctx, args)
	results.QueryRow(func(i int, field sqlcgen.EntryField, err error) {
		resultRow(i, convertEntryFieldModel(field), err)
	})
	if err := results.Close(); err != nil {
		return convertErr(err, "close results for batch create entry fields")
	}
	return nil
}

func (e *EntryFieldRepo) GetFieldsByEntryIDs(ctx context.Context, entryIDs []uuid.UUID) ([]models.EntryField, error) {
	dbFields, err := e.q.EntryFields_GetFieldsByEntryIDs(ctx, entryIDs)
	if err != nil {
		return nil, convertErr(err, "get fields by entry ids %+v", entryIDs)
	}
	var result = make([]models.EntryField, len(dbFields))
	for i, dbField := range dbFields {
		result[i] = *convertEntryFieldModel(dbField)
	}
	return result, nil
}

func convertEntryFieldModel(dbModel sqlcgen.EntryField) *models.EntryField {
	return &models.EntryField{
		BaseModel: &models.BaseModel{
			ID:        dbModel.ID,
			CreatedAt: dbModel.CreatedAt.Time,
			UpdatedAt: dbModel.UpdatedAt.Time,
		},
		EntryID:   dbModel.EntryID,
		Key:       dbModel.Key,
		Value:     dbModel.Value,
		IsPrivate: dbModel.IsPrivate,
	}
}
