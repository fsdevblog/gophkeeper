package models

import (
	"github.com/google/uuid"
)

type EntryType string

const (
	EntryTypeAuth       EntryType = "auth"
	EntryTypeNote       EntryType = "note"
	EntryTypeCreditCard EntryType = "credit_card"
	EntryTypeSSH        EntryType = "ssh"
	EntryTypeDatabase   EntryType = "database"
	EntryTypeServer     EntryType = "server"
)

type Entry struct {
	*BaseModel
	UserID      uuid.UUID
	DeviceID    uuid.UUID
	EntryType   EntryType
	Title       string
	EntryFields []EntryField
}
