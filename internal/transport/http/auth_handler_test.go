package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/fsdevblog/gophkeeper/internal/transport/http/dto"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/mocks"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type AuthHandlerSuite struct {
	suite.Suite
	router    *gin.Engine
	ctrl      *gomock.Controller
	mockAuth  *mocks.MockAuthProvider
	jwtSecret []byte
}

func TestAuthHandlers(t *testing.T) {
	suite.Run(t, new(AuthHandlerSuite))
}

func (s *AuthHandlerSuite) SetupTest() {
	// configure authenticate
	s.ctrl = gomock.NewController(s.T())
	s.mockAuth = mocks.NewMockAuthProvider(s.ctrl)
	s.jwtSecret = []byte("secret")
	s.router = MustNew(InitArgs{
		JWTSecret:    s.jwtSecret,
		AuthProvider: s.mockAuth,
		Logger:       zap.NewNop(),
	})
}

func (s *AuthHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *AuthHandlerSuite) TestLogin() {
	validUser := models.User{
		BaseModel: &models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
		Username:          "valid-user",
		EncryptedPassword: "<ENCRYPTED_PASSWORD>",
	}
	validRequestArgs := dto.AuthenticateParams{
		Username: validUser.Username,
		Password: "<PASSWORD>",
		Device: dto.DeviceParams{
			DeviceType:      models.DeviceTypeCLI,
			Platform:        "ubuntu linux",
			PlatformVersion: gofakeit.AppVersion(),
			AppVersion:      gofakeit.AppVersion(),
		},
	}
	invalidCredsArgs := validRequestArgs
	invalidCredsArgs.Username = "wrong-pass-user"

	s.mockAuth.
		EXPECT().
		Authenticate(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args svcauth.AuthenticateArgs) (string, *models.User, error) {
			switch args.Username {
			case validRequestArgs.Username:
				return "token", &validUser, nil
			case invalidCredsArgs.Username:
				return "", nil, svcauth.ErrInvalidCredentials
			}
			return "", nil, errors.New("")
		}).MinTimes(2)

	tests := []struct {
		name        string
		requestArgs dto.AuthenticateParams
		clientUUID  uuid.UUID
		wantStatus  int
	}{
		{
			name:        "valid request",
			requestArgs: validRequestArgs,
			clientUUID:  uuid.New(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "invalid credentials",
			requestArgs: invalidCredsArgs,
			clientUUID:  uuid.New(),
			wantStatus:  http.StatusUnauthorized,
		}, {
			name:        "with blank device hash header",
			requestArgs: invalidCredsArgs,
			clientUUID:  uuid.Nil,
			wantStatus:  http.StatusBadRequest,
		}, {
			name: "empty device params",
			requestArgs: dto.AuthenticateParams{
				Username: validUser.Username,
				Password: "<PASSWORD>",
			},
			clientUUID: uuid.New(),
			wantStatus: http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var body []byte
			var errEncodeJSON error

			body, errEncodeJSON = json.Marshal(tt.requestArgs)
			s.Require().NoError(errEncodeJSON)

			var options []func(*testutils.RequestOptions)

			if tt.clientUUID != uuid.Nil {
				options = append(options, testutils.WithHeader(middlewares.DeviceHashHeaderKey, tt.clientUUID.String()))
			}

			response, errResponse := testutils.MakeRequest(testutils.RequestArgs{
				Router: s.router,
				Method: http.MethodPost,
				URL:    "/api/auth/login",
				Body:   bytes.NewReader(body),
			}, options...)

			s.Require().NoError(errResponse)
			s.Equal(tt.wantStatus, response.StatusCode)
		})
	}
}

func (s *AuthHandlerSuite) TestRegister() {
	validUser := models.User{
		BaseModel: &models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
		Username: "valid-user",
	}
	validRequestArgs := dto.RegisterParams{
		Username: "valid-user",
		Password: "<PASSWORD>",
		Device: dto.DeviceParams{
			DeviceType:      models.DeviceTypeCLI,
			Platform:        gofakeit.Word(),
			PlatformVersion: gofakeit.AppVersion(),
			AppVersion:      gofakeit.AppVersion(),
		},
	}

	existingUserArgs := validRequestArgs
	existingUserArgs.Username = "existing-user"

	tests := []struct {
		name          string
		requestParams dto.RegisterParams
		wantStatus    int
		clientUUID    uuid.UUID
	}{
		{
			name:          "success",
			requestParams: validRequestArgs,
			clientUUID:    uuid.New(),
			wantStatus:    http.StatusOK,
		}, {
			name:          "user already registered",
			clientUUID:    uuid.New(),
			requestParams: existingUserArgs,
			wantStatus:    http.StatusConflict,
		},
	}

	s.mockAuth.
		EXPECT().
		Register(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args svcauth.RegisterArgs) (string, *models.User, error) {
			switch args.Username {
			case validRequestArgs.Username:
				return "token", &validUser, nil
			case existingUserArgs.Username:
				return "", nil, svcauth.ErrUserAlreadyRegistered
			}
			return "", nil, errors.New("")
		}).MinTimes(2)

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var body []byte
			var errEncodeJSON error

			body, errEncodeJSON = json.Marshal(tt.requestParams)
			s.Require().NoError(errEncodeJSON)

			var options []func(*testutils.RequestOptions)

			if tt.clientUUID != uuid.Nil {
				options = append(options, testutils.WithHeader(middlewares.DeviceHashHeaderKey, tt.clientUUID.String()))
			}
			response, errResponse := testutils.MakeRequest(testutils.RequestArgs{
				Router: s.router,
				Method: http.MethodPost,
				URL:    "/api/auth/register",
				Body:   bytes.NewReader(body),
			}, options...)
			s.Require().NoError(errResponse)
			s.Equal(tt.wantStatus, response.StatusCode)
		})
	}
}
