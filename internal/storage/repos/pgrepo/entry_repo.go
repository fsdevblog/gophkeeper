package pgrepo

import (
	"context"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/sqlcgen"
	"github.com/google/uuid"
)

type EntryRepo struct {
	q *sqlcgen.Queries
}

func NewEntryRepo(conn sqlcgen.DBTX) *EntryRepo {
	return &EntryRepo{q: sqlcgen.New(conn)}
}

func (e *EntryRepo) Create(ctx context.Context, args repodto.CreateEntryArgs) (*models.Entry, error) {
	dbEntry, err := e.q.Entries_Create(ctx, sqlcgen.Entries_CreateParams{
		Type:     args.EntryType,
		UserID:   args.UserID,
		DeviceID: args.DeviceID,
		Title:    args.Title,
	})
	if err != nil {
		return nil, convertErr(err, "create entry")
	}
	return convertEntryModel(dbEntry), nil
}

func (e *EntryRepo) GetAllByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]models.Entry, error) {
	dbEntries, err := e.q.Entries_GetAllByUserID(ctx, sqlcgen.Entries_GetAllByUserIDParams{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, convertErr(err, "get all entries by user id %q", userID)
	}
	var result = make([]models.Entry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		result[i] = *convertEntryModel(dbEntry)
	}
	return result, nil
}

func (e *EntryRepo) GetCountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	totalRecords, err := e.q.Entries_GetCountByUserID(ctx, userID)
	if err != nil {
		return 0, convertErr(err, "get count entries by user id %q", userID)
	}
	return totalRecords, nil
}

func convertEntryModel(dbModel sqlcgen.Entry) *models.Entry {
	return &models.Entry{
		BaseModel: &models.BaseModel{
			ID:        dbModel.ID,
			CreatedAt: dbModel.CreatedAt.Time,
			UpdatedAt: dbModel.UpdatedAt.Time,
		},
		UserID:    dbModel.UserID,
		DeviceID:  dbModel.DeviceID,
		EntryType: dbModel.Type,
		Title:     dbModel.Title,
	}
}
