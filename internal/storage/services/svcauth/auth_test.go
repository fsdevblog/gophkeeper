package svcauth

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"github.com/fsdevblog/gophkeeper/internal/domain"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	repomocks "github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth/mocks"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth/psswd"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	umocks "github.com/fsdevblog/gophkeeper/internal/storage/uow/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type AuthServiceSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockUOW        *umocks.MockUOW
	mockTX         *umocks.MockTX
	mockUserRepo   *repomocks.MockUserRepository
	mockDeviceRepo *repomocks.MockDeviceRepository
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

func (s *AuthServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserRepo = repomocks.NewMockUserRepository(s.ctrl)
	s.mockDeviceRepo = repomocks.NewMockDeviceRepository(s.ctrl)
	s.mockUOW = umocks.NewMockUOW(s.ctrl)
	s.mockTX = umocks.NewMockTX(s.ctrl)

	// configuring mock UOW.
	s.mockUOW.EXPECT().GetRepository(gomock.Any()).
		DoAndReturn(func(repoName uow.RepoName) (uow.Repository, error) {
			return s.repoFactory(repoName)
		}).AnyTimes()

	// configuring mock TX.
	s.mockTX.EXPECT().Get(gomock.Any()).
		DoAndReturn(func(repoName uow.RepoName) (uow.Repository, error) {
			return s.repoFactory(repoName)
		}).AnyTimes()

	// configuring mock UOW.Do().
	s.mockUOW.EXPECT().
		Do(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			fn func(context.Context, uow.TX) error,
			_ ...func(*uow.TransactionOptions),
		) error {
			return fn(ctx, s.mockTX)
		}).AnyTimes()
}

func (s *AuthServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *AuthServiceSuite) TestAuthenticate() {
	validPassword := "<PASSWORD>"
	encryptedPassword, errPass := new(psswd.PasswordHash).HashPassword(validPassword)
	s.Require().NoError(errPass)

	validUser := &models.User{
		Username:          "test",
		EncryptedPassword: encryptedPassword,
	}
	tests := []struct {
		name    string
		args    AuthenticateArgs
		wantErr error
	}{
		{
			name: "success",
			args: AuthenticateArgs{
				Username: validUser.Username,
				Password: validPassword,
				Device: repodto.CreateDeviceArgs{
					DeviceType:      models.DeviceTypeCLI,
					DeviceID:        uuid.New(),
					Platform:        gofakeit.Word(),
					PlatformVersion: gofakeit.AppVersion(),
					AppVersion:      gofakeit.AppVersion(),
				},
			},
			wantErr: nil,
		}, {
			name:    "wrong password",
			args:    AuthenticateArgs{Username: validUser.Username, Password: "<WRONG PASSWORD>"},
			wantErr: ErrInvalidPassword,
		},
	}

	// configuring mock UserRepository.
	s.mockUserRepo.EXPECT().
		FindByUsername(gomock.Any(), validUser.Username).
		Return(validUser, nil).
		MinTimes(2)

	s.mockDeviceRepo.EXPECT().
		Create(gomock.Any(), validUser.ID, gomock.Any()).
		Return(nil).
		MinTimes(1)

	// lets go.
	svc, errSvc := New(s.mockUOW, []byte("secret"))
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NoError(errSvc)
			token, user, err := svc.Authenticate(s.T().Context(), tt.args)
			if tt.wantErr != nil {
				s.Require().Error(err)
				s.ErrorIs(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			s.NotEmpty(token)
			s.NotEmpty(user)
		})
	}
}

func (s *AuthServiceSuite) TestRegister() {
	password := "<PASSWORD>"

	successUser := &models.User{
		Username: "test",
	}
	existingUser := &models.User{
		Username: "existing user",
	}

	s.mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args repodto.CreateUserArgs) (*models.User, error) {
			if args.Username == successUser.Username {
				return successUser, nil
			}
			return nil, domain.ErrDuplicateKey
		}).MinTimes(2)

	s.mockDeviceRepo.EXPECT().
		Create(gomock.Any(), successUser.ID, gomock.Any()).
		Return(nil).
		Times(1)

	tests := []struct {
		name    string
		args    RegisterArgs
		wantErr error
	}{
		{
			name: "success",
			args: RegisterArgs{
				Username: successUser.Username,
				Password: password,
				Device: repodto.CreateDeviceArgs{
					DeviceType:      models.DeviceTypeCLI,
					DeviceID:        uuid.New(),
					Platform:        gofakeit.Word(),
					PlatformVersion: gofakeit.AppVersion(),
					AppVersion:      gofakeit.AppVersion(),
				},
			},
			wantErr: nil,
		}, {
			name:    "existing user",
			args:    RegisterArgs{Username: existingUser.Username, Password: password},
			wantErr: ErrUserAlreadyRegistered,
		},
	}

	svc, errSvc := New(s.mockUOW, []byte("secret"))
	s.Require().NoError(errSvc)

	for _, tt := range tests {
		s.Run(tt.name, func() {
			token, user, err := svc.Register(s.T().Context(), tt.args)
			if tt.wantErr != nil {
				s.Require().Error(err)
				s.ErrorIs(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			s.NotEmpty(token)
			s.NotEmpty(user)
		})
	}
}

func (s *AuthServiceSuite) repoFactory(repoName uow.RepoName) (uow.Repository, error) {
	switch repoName {
	case uow.RepoName(repodto.UserRepoName):
		return s.mockUserRepo, nil
	case uow.RepoName(repodto.DeviceRepoName):
		return s.mockDeviceRepo, nil
	}
	return nil, fmt.Errorf("unknown repository: %s", repoName)
}
