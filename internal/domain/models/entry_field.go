package models

import "github.com/google/uuid"

type EntryFieldKeyType string

const (
	FieldKeyUsername EntryFieldKeyType = "username"
	FieldKeyEmail    EntryFieldKeyType = "email"
	FieldKeyPassword EntryFieldKeyType = "password"
	FieldKeyOTP      EntryFieldKeyType = "otp"
	FieldKeyURL      EntryFieldKeyType = "url"
)

type EntryField struct {
	*BaseModel
	EntryID   uuid.UUID
	Key       string
	Value     []byte
	IsPrivate bool
}
