package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceType string

const (
	DeviceTypeCLI     DeviceType = "cli"
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeMobile  DeviceType = "mobile"
	DeviceTypeWeb     DeviceType = "web"
)

type Device struct {
	*BaseModel
	UserID          uuid.UUID
	DeviceType      DeviceType
	DeviceID        uuid.UUID // Uniq device's identifier. Must generate on client side.
	Platform        string
	PlatformVersion string
	StateVersion    uuid.UUID // Current state version of synchronization.
	AppVersion      string
	LastActiveAt    time.Time
}
