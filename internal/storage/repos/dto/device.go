package dto

import (
	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/google/uuid"
)

type CreateDeviceArgs struct {
	DeviceType      models.DeviceType
	DeviceHash      uuid.UUID
	Platform        string
	PlatformVersion string
	AppVersion      string
}
