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
			a.l.Error("login request", zap.Error(errResp))
			return nil, ErrUnknown
		}
		a.l.Error("login request", zap.Error(err))
		return nil, ErrUnknown
	}
	user := &models.User{
		ID:       resp.ID,
		Username: resp.Username,
	}

	if errAuthenticate := a.authenticateUser(user, token); errAuthenticate != nil {
		return nil, errAuthenticate
	}

	return user, nil
}

func (a *Auth) Register(ctx context.Context, username string, password string) (*models.User, error) {
	resp, token, err := a.api.Register(ctx, api.RegisterParams{
		Username: username,
		Password: password,
	})
	if err != nil {
		var errResp *api.UnexpectedHTTPStatusCodeError
		if errors.As(err, &errResp) {
			switch errResp.StatusCode {
			case http.StatusConflict:
				return nil, ErrUserAlreadyExists
			case http.StatusUnprocessableEntity:
				return nil, ErrIncorrectFieldsFormat
			case http.StatusUnauthorized:
				return nil, ErrAlreadyLoggedIn
			}
			a.l.Error("register request", zap.Error(errResp))
			return nil, ErrUnknown
		}
		a.l.Error("register request", zap.Error(err))
		return nil, ErrUnknown
	}

	user := &models.User{
		ID:       resp.ID,
		Username: resp.Username,
	}

	if errAuthenticate := a.authenticateUser(user, token); errAuthenticate != nil {
		return nil, errAuthenticate
	}

	return user, nil
}

func (a *Auth) authenticateUser(user *models.User, token string) error {
	if errSaveToken := keyring.Set(ServiceName, user.Username, token); errSaveToken != nil {
		a.l.Error("saving token", zap.Error(errSaveToken))
		return ErrUnknown
	}
	return nil
}
