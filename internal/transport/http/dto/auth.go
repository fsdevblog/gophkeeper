package dto

import (
	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/google/uuid"
)

type RegisterParams struct {
	Username string       `binding:"required,min=1,max=15"       json:"username"`
	Password string       `binding:"required,min=6,max_bytes=72" json:"password"`
	Device   DeviceParams `binding:"required"                    json:"device"`
}

// DeviceParams fields for device registration. DeviceHash must present in each HTTP header.
type DeviceParams struct {
	DeviceType      models.DeviceType `binding:"required"        json:"device_type"`
	Platform        string            `binding:"required,max=32" json:"platform"`
	PlatformVersion string            `binding:"required,max=16" json:"platform_version"`
	AppVersion      string            `binding:"required,max=16" json:"app_version"`
}

type AuthenticateParams struct {
	Username string       `binding:"required,min=1,max=15"       json:"username"`
	Password string       `binding:"required,min=6,max_bytes=72" json:"password"`
	Device   DeviceParams `binding:"required"                    json:"device"`
}

type UserItem struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}
