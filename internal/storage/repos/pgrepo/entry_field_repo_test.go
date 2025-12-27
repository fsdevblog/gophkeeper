package pgrepo

import (
	"testing"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo/dbtest"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type EntryFieldRepoSuite struct {
	suite.Suite
	conn        *pgxpool.Pool
	repo        *EntryFieldRepo
	pgContainer testcontainers.Container
	fixtures    *testfixtures.Loader
}

func TestEntryFieldRepoSuite(t *testing.T) {
	suite.Run(t, new(EntryRepoSuite))
}

func (s *EntryFieldRepoSuite) SetupSuite() {
	pg, errPg := dbtest.Connect(s.T().Context())
	s.Require().NoError(errPg)

	s.pgContainer = pg.PgContainer
	s.conn = pg.PgPool
	s.fixtures = pg.Fixtures

	s.repo = NewEntryFieldRepo(s.conn)
}

func (s *EntryFieldRepoSuite) SetupTest() {
	err := s.fixtures.Load()
	s.Require().NoError(err)
}

func (s *EntryFieldRepoSuite) TearDownTest() {
	err := dbtest.ClearTables(s.conn, "entries")
	s.Require().NoError(err)
}

func (s *EntryFieldRepoSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(s.T().Context())
		s.Require().NoError(err)
	}
}

func (s *EntryFieldRepoSuite) TestCreate() {
	entryID := uuid.MustParse("1f94b403-a698-4098-b806-8896c9f37b60")
	var fields = []dto.CreateEntryFieldArgs{
		{
			EntryID:   entryID,
			Key:       models.FieldKeyUsername,
			Value:     []byte("username"),
			IsPrivate: false,
		}, {
			EntryID:   entryID,
			Key:       models.FieldKeyPassword,
			Value:     []byte("<PASSWORD>"),
			IsPrivate: true,
		},
	}

	err := s.repo.BatchCreate(s.T().Context(), fields, func(i int, field *models.EntryField, err error) {
		s.Require().NoError(err)
		s.Require().NotNil(field)
		s.Equal(fields[i].Key, field.Key)
		s.Equal(fields[i].Value, field.Value)
		s.Equal(fields[i].IsPrivate, field.IsPrivate)
		s.Equal(entryID, field.EntryID)
	})
	s.Require().NoError(err)
}

func (s *EntryFieldRepoSuite) TestGetFieldsByEntryIDs() {
	entryIDs := []uuid.UUID{
		uuid.MustParse("1f94b403-a698-4098-b806-8896c9f37b60"),
		uuid.MustParse("870ca9c1-bf1a-4731-85d5-2b8b9acba8b3"),
	}
	fields, err := s.repo.GetFieldsByEntryIDs(s.T().Context(), entryIDs)
	s.Require().NoError(err)
	s.Len(fields, 2)
}
