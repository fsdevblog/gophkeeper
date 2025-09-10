package pgrepo

import (
	"testing"

	"github.com/fsdevblog/gophkeeper/internal/storage/repos"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/testutils"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type UserRepoSuite struct {
	suite.Suite
	conn        *pgxpool.Pool
	repo        *UserRepo
	pgContainer testcontainers.Container
	fixtures    *testfixtures.Loader
}

func TestUserRepo(t *testing.T) {
	suite.Run(t, new(UserRepoSuite))
}

func (s *UserRepoSuite) SetupSuite() {
	pg, errPg := testutils.Connect(s.T().Context())
	s.Require().NoError(errPg)

	s.pgContainer = pg.PgContainer
	s.conn = pg.PgPool
	s.fixtures = pg.Fixtures

	s.repo = NewUserRepo(s.conn)
}

func (s *UserRepoSuite) SetupTest() {
	err := s.fixtures.Load()
	s.Require().NoError(err)
}

func (s *UserRepoSuite) TearDownTest() {
	err := testutils.ClearTables(s.conn, "users")
	s.Require().NoError(err)
}

func (s *UserRepoSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(s.T().Context())
		s.Require().NoError(err)
	}
}

func (s *UserRepoSuite) TestFindByUsername() {
	tests := []struct {
		name       string
		username   string
		wantUserID uuid.UUID
		wantErr    error
	}{
		{
			name:       "existing user",
			username:   "first-username",
			wantUserID: uuid.MustParse("5aa9061c-c62e-4588-8dac-2639a4f42349"),
		},
		{
			name:     "unexisting user",
			username: "unexisted",
			wantErr:  repos.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := s.repo.FindByUsername(s.T().Context(), tt.username)
			if tt.wantErr != nil {
				s.Require().Error(err)
				s.ErrorIs(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tt.wantUserID, got.ID)
		})
	}
}

func (s *UserRepoSuite) TestCreate() {
	tests := []struct {
		name    string
		args    repodto.CreateUserArgs
		wantErr error
	}{
		{
			name: "success",
			args: repodto.CreateUserArgs{
				Username:          "new-username",
				EncryptedPassword: "<PASSWORD>",
			},
			wantErr: nil,
		}, {
			name: "existing user",
			args: repodto.CreateUserArgs{
				Username:          "first-username",
				EncryptedPassword: "<PASSWORD>",
			},
			wantErr: repos.ErrDuplicateKey,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := s.repo.Create(s.T().Context(), tt.args)
			if tt.wantErr != nil {
				s.Require().Error(err)
				s.ErrorIs(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tt.args.Username, got.Username)
			s.Equal(tt.args.EncryptedPassword, got.EncryptedPassword)
		})
	}
}
