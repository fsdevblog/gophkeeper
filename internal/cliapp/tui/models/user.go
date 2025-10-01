package models

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Username string
}

type Session struct {
	User            *User
	IsAuthenticated bool
}
