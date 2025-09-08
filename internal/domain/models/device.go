package models

import "github.com/google/uuid"

type Device struct {
	*BaseModel
	UserID uuid.UUID
}
