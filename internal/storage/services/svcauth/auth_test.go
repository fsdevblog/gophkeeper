package svcauth

import (
	"testing"

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
	ctrl         *gomock.Controller
	mockUOW      *umocks.MockUOW
	mockUserRepo *repomocks.MockUserRepository
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

func (s *AuthServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserRepo = repomocks.NewMockUserRepository(s.ctrl)
	s.mockUOW = umocks.NewMockUOW(s.ctrl)

	// configuring mock UOW.
	s.mockUOW.EXPECT().GetRepository(uow.RepoName(repodto.UserRepoName)).
		Return(s.mockUserRepo, nil).AnyTimes()
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
			name:    "success",
			args:    AuthenticateArgs{Username: validUser.Username, Password: validPassword},
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
			s.NotEmpty(token)
			s.NotEmpty(user)
		})
	}
}
