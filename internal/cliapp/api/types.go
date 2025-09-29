package api

import "github.com/google/uuid"

type DeviceRequest struct {
	DeviceType      string `json:"device_type"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	AppVersion      string `json:"app_version"`
}
type ErrResponse struct {
	Error string `json:"error"`
}

type LoginRequest struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Device   DeviceRequest `json:"device"`
}

type LoginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type LoginParams struct {
	Username string
	Password string
}

type RegisterParams struct {
	Username string
	Password string
}

type RegisterResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type RegisterRequest struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Device   DeviceRequest `json:"device"`
}
