package pgrepo

import (
	"context"

	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/sqlcgen"
)

type UserRepo struct {
	q *sqlcgen.Queries
}

func NewUserRepo(conn sqlcgen.DBTX) *UserRepo {
	return &UserRepo{q: sqlcgen.New(conn)}
}

func (u *UserRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	dbUser, err := u.q.Users_FindByUsername(ctx, username)
	if err != nil {
		return nil, convertErr(err, "find user by username %q", username)
	}
	return convertUserModel(dbUser), nil
}

func (u *UserRepo) Create(ctx context.Context, args repodto.CreateUserArgs) (*models.User, error) {
	dbUser, err := u.q.Users_Create(ctx, sqlcgen.Users_CreateParams{
		Username:          args.Username,
		EncryptedPassword: args.EncryptedPassword,
	})
	if err != nil {
		return nil, convertErr(err, "create user")
	}
	return convertUserModel(dbUser), nil
}

func convertUserModel(dbModel sqlcgen.User) *models.User {
	return &models.User{
		BaseModel: &models.BaseModel{
			ID:        dbModel.ID,
			CreatedAt: dbModel.CreatedAt.Time,
			UpdatedAt: dbModel.UpdatedAt.Time,
		},
		Username:          dbModel.Username,
		EncryptedPassword: dbModel.EncryptedPassword,
	}
}
