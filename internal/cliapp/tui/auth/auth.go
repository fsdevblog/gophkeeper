package auth

import (
	"context"
	"errors"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/models"
	"github.com/zalando/go-keyring"
	"go.uber.org/zap"
	"net/http"
)

const ServiceName = "gophkeeper"

type Auth struct {
	api *api.Client
	l   *zap.Logger
}

func New(apiClient *api.Client, l *zap.Logger) *Auth {
	return &Auth{
		api: apiClient,
		l:   l,
	}
}

func (a *Auth) Login(ctx context.Context, username, password string) (*models.User, error) {
	resp, token, err := a.api.Login(ctx, api.LoginParams{
		Username: username,
		Password: password,
	})
	if err != nil {
		var errResp *api.UnexpectedHTTPStatusCodeError
		if errors.As(err, &errResp) {
			switch errResp.StatusCode {
			case http.StatusUnauthorized:
				return nil, ErrLoginCredentialsWrong
			case http.StatusUnprocessableEntity:
				return nil, ErrIncorrectFieldsFormat
			}
			a.l.Error("login request", zap.Error(err))
			return nil, ErrUnknown
		}
		a.l.Error("login request", zap.Error(err))
		return nil, ErrUnknown
	}
	if errSaveToken := keyring.Set(ServiceName, resp.Username, token); errSaveToken != nil {
		a.l.Error("saving token", zap.Error(errSaveToken))
		return nil, ErrUnknown
	}

	return &models.User{
		ID:       resp.ID,
		Username: resp.Username,
	}, nil
}
