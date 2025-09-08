package dto

import (
	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/google/uuid"
)

type CreateEntryArgs struct {
	Title     string
	UserID    uuid.UUID
	DeviceID  uuid.UUID
	EntryType models.EntryType
}

type CreateEntryFieldArgs struct {
	EntryID uuid.UUID
	Key     models.EntryFieldKeyType
	Value   []byte
}
