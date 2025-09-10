package models

type User struct {
	*BaseModel
	Username          string
	EncryptedPassword string
}
