package pgrepo

import (
	"context"

	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/sqlcgen"
	"github.com/google/uuid"
)

type DeviceRepo struct {
	q *sqlcgen.Queries
}

func NewDeviceRepo(conn sqlcgen.DBTX) *DeviceRepo {
	return &DeviceRepo{q: sqlcgen.New(conn)}
}

func (d *DeviceRepo) Create(ctx context.Context, userID uuid.UUID, args repodto.CreateDeviceArgs) error {
	_, err := d.q.Devices_Create(ctx, sqlcgen.Devices_CreateParams{
		Type:            args.DeviceType,
		UserID:          userID,
		ClientUuid:      args.DeviceHash,
		Platform:        args.Platform,
		PlatformVersion: args.PlatformVersion,
		StateVersion:    args.StateVersion,
		AppVersion:      args.AppVersion,
	})
	if err != nil {
		return convertErr(err, "create device")
	}
	return nil
}
