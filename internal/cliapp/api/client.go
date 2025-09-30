package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fsdevblog/gophkeeper/internal/cliapp/device"
	"resty.dev/v3"
)

type Client struct {
	r      *resty.Client
	device *device.Device
}

func New(baseURL string) (*Client, error) {
	d, err := device.New()
	if err != nil {
		return nil, fmt.Errorf("init http client: %w", err)
	}

	r := resty.New().
		SetBaseURL(baseURL)

	return &Client{
		r:      r,
		device: d,
	}, nil
}

func (c *Client) Login(ctx context.Context, params LoginParams) (*LoginResponse, string, error) {
	var resp LoginResponse
	var errResp ErrResponse

	result, err := c.r.R().
		SetContext(ctx).
		SetBody(LoginRequest{
			Username: params.Username,
			Password: params.Password,
			Device: DeviceRequest{
				DeviceType:      c.device.DeviceType,
				Platform:        c.device.PlatformName,
				PlatformVersion: c.device.PlatformVersion,
				AppVersion:      "0.0.1",
			},
		}).
		SetResult(&resp).
		SetHeader("X-Device-Hash", c.device.UUID.String()).
		SetError(&errResp).
		Post("/api/auth/login")
	if err != nil {
		return nil, "", fmt.Errorf("authenticate request: %w", err)
	}
	if result.IsError() {
		return nil, "", NewUnexpectedHTTPStatusCodeError(result.StatusCode(), errResp.Error)
	}

	tokenHeader := result.Header().Get("Authorization")
	if !strings.HasPrefix(tokenHeader, "Bearer ") {
		return nil, "", errors.New("invalid token header")
	}
	token := tokenHeader[7:]

	return &resp, token, nil
}

func (c *Client) Register(ctx context.Context, params RegisterParams) (*RegisterResponse, string, error) {
	var resp RegisterResponse
	var errResp ErrResponse
	result, err := c.r.
		R().
		SetContext(ctx).
		SetBody(RegisterRequest{
			Username: params.Username,
			Password: params.Password,
			Device: DeviceRequest{
				DeviceType:      c.device.DeviceType,
				Platform:        c.device.PlatformName,
				PlatformVersion: c.device.PlatformVersion,
				AppVersion:      "0.0.1",
			},
		}).
		SetResult(&resp).
		SetHeader("X-Device-Hash", c.device.UUID.String()).
		SetError(&errResp).
		Post("/api/auth/register")

	if err != nil {
		return nil, "", fmt.Errorf("register request: %w", err)
	}
	if result.IsError() {
		return nil, "", NewUnexpectedHTTPStatusCodeError(result.StatusCode(), errResp.Error)
	}

	tokenHeader := result.Header().Get("Authorization")
	if !strings.HasPrefix(tokenHeader, "Bearer ") {
		return nil, "", errors.New("invalid token header")
	}
	token := tokenHeader[7:]

	return &resp, token, nil
}
