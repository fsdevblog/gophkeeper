package dto

import (
	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/google/uuid"
)

type EntryResponse struct {
	Entries []EntryResponseItem `json:"entries"`
	Meta    PaginationMeta      `json:"meta"`
}
type EntryResponseItem struct {
	ID        uuid.UUID        `json:"id"`
	Title     string           `json:"title"`
	EntryType models.EntryType `json:"entryType"`
}

type EntryFieldResponseItem struct {
	ID        uuid.UUID                `json:"id"`
	EntryID   uuid.UUID                `json:"entryId"`
	Key       models.EntryFieldKeyType `json:"key"`
	Value     []byte                   `json:"value,omitempty"`
	IsPrivate bool                     `json:"isPrivate"`
}
