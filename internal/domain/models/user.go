package models

import (
	"github.com/google/uuid"
)

type User struct {
	*BaseModel
	ID                uuid.UUID
	Username          string
	EncryptedPassword string
}
