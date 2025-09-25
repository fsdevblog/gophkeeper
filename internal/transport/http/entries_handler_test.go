package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/transport/http/dto"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	"github.com/fsdevblog/gophkeeper/internal/tokens"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/mocks"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/stretchr/testify/suite"
)

type EntriesHandlerSuite struct {
	suite.Suite
	router    *gin.Engine
	ctrl      *gomock.Controller
	mockEntry *mocks.MockEntryProvider
	jwtSecret []byte
}

func TestEntriesHandlerSuite(t *testing.T) {
	suite.Run(t, new(EntriesHandlerSuite))
}

func (s *EntriesHandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockEntry = mocks.NewMockEntryProvider(s.ctrl)
	s.jwtSecret = []byte("secret")
	s.router = MustNew(InitArgs{
		JWTSecret:     s.jwtSecret,
		EntryProvider: s.mockEntry,
		Logger:        zap.NewNop(),
	})
}

func (s *EntriesHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *EntriesHandlerSuite) Test_GetAll() {
	userUUID := uuid.New()
	clientUUID := uuid.New()
	results := []models.Entry{
		{
			BaseModel: &models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID:    userUUID,
			DeviceID:  clientUUID,
			EntryType: models.EntryTypeSSH,
			Title:     "ssh access",
		}, {
			BaseModel: &models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID:    userUUID,
			DeviceID:  uuid.New(),
			EntryType: models.EntryTypeNote,
			Title:     "some note",
		},
	}

	var expectedEntries = make([]dto.EntryResponseItem, len(results))
	for i, entry := range results {
		expectedEntries[i] = dto.EntryResponseItem{
			ID:        entry.ID,
			Title:     entry.Title,
			EntryType: entry.EntryType,
		}
	}

	s.mockEntry.
		EXPECT().
		GetUserEntries(gomock.Any(), userUUID, gomock.Any()).
		Return(results, int64(len(results)), nil).
		Times(1)

	userToken, errToken := tokens.GenerateUserJWT(userUUID, time.Hour, s.jwtSecret)
	s.Require().NoError(errToken)

	response, errResponse := testutils.MakeRequest(testutils.RequestArgs{
		Router: s.router,
		Method: http.MethodGet,
		URL:    "/api/entries",
	},
		testutils.WithHeader(middlewares.DeviceHashHeaderKey, clientUUID.String()),
		testutils.WithHeader("Authorization", fmt.Sprintf("Bearer %s", userToken)),
	)

	s.Require().NoError(errResponse)
	defer response.Body.Close()

	s.Equal(http.StatusOK, response.StatusCode)

	bodyBytes, errRead := io.ReadAll(response.Body)
	s.Require().NoError(errRead)

	var responseJSON dto.EntryResponse
	errUnmarshal := json.Unmarshal(bodyBytes, &responseJSON)
	s.Require().NoError(errUnmarshal)
	s.Equal(expectedEntries, responseJSON.Entries)
	s.Equal(int64(len(results)), responseJSON.Meta.TotalRecords)
}

func (s *EntriesHandlerSuite) Test_GetEntryFields() {
	userUUID := uuid.New()
	entryUUID := uuid.New()
	clientUUID := uuid.New()

	fields := []models.EntryField{
		{
			BaseModel: &models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			EntryID:   entryUUID,
			Key:       models.FieldKeyUsername,
			Value:     []byte("username"),
			IsPrivate: false,
		}, {
			BaseModel: &models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			EntryID:   entryUUID,
			Key:       models.FieldKeyPassword,
			Value:     nil, // Private field should not be returned.
			IsPrivate: true,
		},
	}

	var expectedResponse = make([]dto.EntryFieldResponseItem, len(fields))
	for i, field := range fields {
		expectedResponse[i] = dto.EntryFieldResponseItem{
			ID:        field.ID,
			EntryID:   field.EntryID,
			Key:       field.Key,
			Value:     field.Value,
			IsPrivate: field.IsPrivate,
		}
	}

	s.mockEntry.
		EXPECT().
		GetSafeEntryFields(gomock.Any(), userUUID, entryUUID).
		Return(fields, nil).
		Times(1)

	userToken, errToken := tokens.GenerateUserJWT(userUUID, time.Hour, s.jwtSecret)
	s.Require().NoError(errToken)

	response, errResponse := testutils.MakeRequest(testutils.RequestArgs{
		Router: s.router,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/entries/%s/fields", entryUUID.String()),
	},
		testutils.WithHeader(middlewares.DeviceHashHeaderKey, clientUUID.String()),
		testutils.WithHeader("Authorization", fmt.Sprintf("Bearer %s", userToken)),
	)
	s.Require().NoError(errResponse)
	defer response.Body.Close()
	s.Equal(http.StatusOK, response.StatusCode)

	bodyBytes, errRead := io.ReadAll(response.Body)
	s.Require().NoError(errRead)

	var responseJSON []dto.EntryFieldResponseItem
	errUnmarshal := json.Unmarshal(bodyBytes, &responseJSON)
	s.Require().NoError(errUnmarshal)
	s.Equal(expectedResponse, responseJSON)
}
