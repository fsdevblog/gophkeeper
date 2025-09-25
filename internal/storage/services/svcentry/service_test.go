package svcentry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/pag"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	emocks "github.com/fsdevblog/gophkeeper/internal/storage/services/svcentry/mocks"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	umocks "github.com/fsdevblog/gophkeeper/internal/storage/uow/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type EntryServiceSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	mockUOW            *umocks.MockUOW
	mockTX             *umocks.MockTX
	mockEntryRepo      *emocks.MockEntryRepository
	mockEntryFieldRepo *emocks.MockEntryFieldRepository
}

func TestEntryService(t *testing.T) {
	suite.Run(t, new(EntryServiceSuite))
}

func (s *EntryServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUOW = umocks.NewMockUOW(s.ctrl)
	s.mockTX = umocks.NewMockTX(s.ctrl)
	s.mockEntryRepo = emocks.NewMockEntryRepository(s.ctrl)
	s.mockEntryFieldRepo = emocks.NewMockEntryFieldRepository(s.ctrl)

	// configuring mock UOW.
	s.mockUOW.EXPECT().
		GetRepository(gomock.Any()).
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

	// configuring mock UOW.Do().
	s.mockUOW.EXPECT().DoWithIsolation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, fn func(context.Context, uow.TX) error, _ uow.IsolationLevel, _ uint) error {
			return fn(s.T().Context(), s.mockTX)
		},
	).AnyTimes()
}

func (s *EntryServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *EntryServiceSuite) TestCreateEntry() {
	s.Run("success", func() {
		testingArgs := CreateEntryArgs{
			Title:     "test.com",
			UserID:    uuid.New(),
			DeviceID:  uuid.New(),
			EntryType: models.EntryTypeAuth,
			EntryFields: []EntryFieldArg{
				{
					Key:       models.FieldKeyUsername,
					Value:     []byte("test"),
					IsPrivate: false,
				}, {
					Key:       models.FieldKeyPassword,
					Value:     []byte("strong_password"),
					IsPrivate: true,
				}, {
					Key:       models.FieldKeyURL,
					Value:     []byte("https://auth.test.com"),
					IsPrivate: false,
				}, {
					Key:       models.FieldKeyOTP,
					Value:     []byte("blablabla"),
					IsPrivate: true,
				},
			},
		}

		createdEntryID := uuid.New()
		// configure mock EntryRepository.
		s.mockEntryRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, args repodto.CreateEntryArgs) (*models.Entry, error) {
				s.Equal(testingArgs.Title, args.Title)
				s.Equal(testingArgs.UserID, args.UserID)
				s.Equal(testingArgs.DeviceID, args.DeviceID)
				s.Equal(testingArgs.EntryType, args.EntryType)

				return &models.Entry{
					BaseModel: &models.BaseModel{
						ID:        createdEntryID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					UserID:    args.UserID,
					DeviceID:  args.DeviceID,
					EntryType: args.EntryType,
					Title:     args.Title,
				}, nil
			}).Times(1)

		// configure mock of EntryFieldRepository.
		s.mockEntryFieldRepo.EXPECT().
			BatchCreate(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				_ context.Context,
				fields []repodto.CreateEntryFieldArgs,
				_ func(i int, field *models.EntryField, err error),
			) error {
				s.Equal(createdEntryID, fields[0].EntryID)
				s.Len(fields, len(testingArgs.EntryFields))
				return nil
			}).Times(1)

		// lets goooo

		svc := New(s.mockUOW)
		entry, err := svc.CreateEntry(s.T().Context(), testingArgs)
		s.Require().NoError(err)
		s.NotEmpty(entry)
		s.NotEmpty(entry.EntryFields)
		s.Len(entry.EntryFields, len(testingArgs.EntryFields))
	})
}

func (s *EntryServiceSuite) TestGetUserEntries() {
	testingUserUUID := uuid.New()
	userEntryID := uuid.New()
	pagination := pag.New()
	wantEntryFields := []models.EntryField{
		{
			BaseModel: &models.BaseModel{
				ID: uuid.New(),
			},
			EntryID:   userEntryID,
			Key:       "login",
			Value:     []byte("test"),
			IsPrivate: false,
		}, {
			BaseModel: &models.BaseModel{
				ID: uuid.New(),
			},
			EntryID:   userEntryID,
			Key:       "password",
			Value:     []byte("<PASSWORD>"),
			IsPrivate: true,
		}, {
			BaseModel: &models.BaseModel{
				ID: uuid.New(),
			},
			EntryID:   userEntryID,
			Key:       "url",
			Value:     []byte("https://auth.test.com"),
			IsPrivate: false,
		},
	}
	wantEntries := []models.Entry{
		{
			BaseModel: &models.BaseModel{
				ID: userEntryID,
			},
			UserID:    testingUserUUID,
			DeviceID:  uuid.New(),
			EntryType: models.EntryTypeAuth,
			Title:     "test.com",
		},
	}
	svc := New(s.mockUOW)
	// configure mock EntryRepository.
	s.mockEntryRepo.EXPECT().
		GetAllByUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			_ context.Context,
			userID uuid.UUID,
			limit int32,
			offset int32,
		) ([]models.Entry, int, error) {
			// checking entire arguments.
			s.Equal(testingUserUUID, userID)
			s.Equal(pagination.Limit(), limit)
			s.Equal(pagination.Offset(), offset)
			return wantEntries, len(wantEntries), nil
		}).
		Times(1)

	s.mockEntryRepo.EXPECT().
		GetCountByUser(gomock.Any(), testingUserUUID).
		Return(int64(len(wantEntries)), nil).
		Times(1)

	// configure mock EntryFieldRepository.
	s.mockEntryFieldRepo.EXPECT().
		GetFieldsByEntryIDs(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, entryIDs []uuid.UUID) ([]models.EntryField, error) {
			var expectingEntryIDs = make([]uuid.UUID, len(wantEntries))
			for i, entry := range wantEntries {
				expectingEntryIDs[i] = entry.ID
			}
			// checking entryIDs, it must be equal to the IDs of entries we got from the mock EntryRepository.GetAllByUser().
			s.Equal(expectingEntryIDs, entryIDs)
			return wantEntryFields, nil
		}).
		Times(1)

	// lets go testing!!
	entries, totalRecords, err := svc.GetUserEntries(s.T().Context(), testingUserUUID, pagination)
	s.Require().NoError(err)
	s.Equal(int64(len(wantEntries)), totalRecords)
	s.Equal(wantEntryFields, entries[0].EntryFields)
}

func (s *EntryServiceSuite) TestGetSafeEntryFields() {
	entryID := uuid.New()
	userID := uuid.New()
	fields := []models.EntryField{
		{
			BaseModel: &models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			EntryID:   entryID,
			Key:       models.FieldKeyUsername,
			Value:     []byte("test"),
			IsPrivate: false,
		}, {
			BaseModel: &models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			EntryID:   entryID,
			Key:       models.FieldKeyPassword,
			Value:     []byte("password"),
			IsPrivate: true,
		},
	}
	s.mockEntryFieldRepo.
		EXPECT().
		GetUserFieldsByEntryID(gomock.Any(), entryID, userID).
		Return(fields, nil).
		Times(1)

	svc := New(s.mockUOW)
	gotFields, err := svc.GetSafeEntryFields(s.T().Context(), userID, entryID)
	s.Require().NoError(err)
	for _, field := range gotFields {
		if field.IsPrivate {
			s.Nil(field.Value)
			continue
		}
		s.NotNil(field.Value)
	}
}

func (s *EntryServiceSuite) repoFactory(repoName uow.RepoName) (uow.Repository, error) {
	switch repoName {
	case uow.RepoName(repodto.EntryRepoName):
		return s.mockEntryRepo, nil
	case uow.RepoName(repodto.EntryFieldRepoName):
		return s.mockEntryFieldRepo, nil
	}
	return nil, fmt.Errorf("unknown repository: %s", repoName)
}
