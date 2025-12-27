package pgrepo

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/dbtest"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type EntryRepoSuite struct {
	suite.Suite
	conn        *pgxpool.Pool
	repo        *EntryRepo
	pgContainer testcontainers.Container
	fixtures    *testfixtures.Loader
}

func TestEntryRepoSuite(t *testing.T) {
	suite.Run(t, new(EntryRepoSuite))
}

func (s *EntryRepoSuite) SetupSuite() {
	pg, errPg := dbtest.Connect(s.T().Context())
	s.Require().NoError(errPg)

	s.pgContainer = pg.PgContainer
	s.conn = pg.PgPool
	s.fixtures = pg.Fixtures

	s.repo = NewEntryRepo(s.conn)
}

func (s *EntryRepoSuite) SetupTest() {
	err := s.fixtures.Load()
	s.Require().NoError(err)
}

func (s *EntryRepoSuite) TearDownTest() {
	err := dbtest.ClearTables(s.conn, "entries")
	s.Require().NoError(err)
}

func (s *EntryRepoSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(s.T().Context())
		s.Require().NoError(err)
	}
}

func (s *EntryRepoSuite) TestCreate() {
	entry, err := s.repo.Create(s.T().Context(), repodto.CreateEntryArgs{
		Title:     gofakeit.Sentence(3),
		UserID:    uuid.MustParse("5aa9061c-c62e-4588-8dac-2639a4f42349"),
		DeviceID:  uuid.MustParse("229e6592-befe-4dad-bbe6-4e67fdea5e57"),
		EntryType: models.EntryTypeAuth,
	})
	s.Require().NoError(err)
	s.NotNil(entry)
}

func (s *EntryRepoSuite) TestGetAllByUser() {
	userID := uuid.MustParse("5aa9061c-c62e-4588-8dac-2639a4f42349")
	entries, err := s.repo.GetAllByUser(s.T().Context(), userID, 10, 0)
	s.Require().NoError(err)
	s.Len(entries, 2) // we have 2 records into fixtures
}

func (s *EntryRepoSuite) TestGetCountByUser() {
	userID := uuid.MustParse("5aa9061c-c62e-4588-8dac-2639a4f42349")
	count, err := s.repo.GetCountByUser(s.T().Context(), userID)
	s.Require().NoError(err)
	s.Equal(int64(2), count)
}
