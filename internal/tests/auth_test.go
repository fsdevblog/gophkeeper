package tests

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/fsdevblog/gophkeeper/internal/db/dbtest"
	"github.com/fsdevblog/gophkeeper/internal/storage/services"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	apphttp "github.com/fsdevblog/gophkeeper/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
	"net/http"
	"resty.dev/v3"
	"testing"
	"time"
)

const (
	defaultTimeout = 10 * time.Second
)

type AuthIntegrationSuite struct {
	suite.Suite
	authService *svcauth.AuthService
	pgContainer testcontainers.Container
	baseURL     string
	conn        *pgxpool.Pool
}

func TestAuthIntegration(t *testing.T) {
	suite.Run(t, new(AuthIntegrationSuite))
}

func (s *AuthIntegrationSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(s.T().Context(), defaultTimeout)
	defer cancel()

	pg, errPg := dbtest.Connect(ctx, func(c *dbtest.ConnectionConfig) {
		c.DatabaseName = "gophkeeper_test"
		c.FixturesPath = "testdata/fixtures"
	})
	s.Require().NoError(errPg)
	s.baseURL = "http://127.0.0.1:8080"
	s.pgContainer = pg.PgContainer
	s.conn = pg.PgPool

	unitOfWork := uow.New(s.conn)

	// auth service init
	jwtSecret := []byte(gofakeit.UUID())
	s.authService = svcauth.New(unitOfWork, jwtSecret)

	// router init
	router, errRouter := apphttp.New(apphttp.InitArgs{
		JWTSecret: jwtSecret,
		Services: &services.Collection{
			AuthService: s.authService,
		},
		Logger: zap.NewExample(),
	})
	s.Require().NoError(errRouter)

	go func() {
		mustStartHTTPServer(s.T().Context(), ":8080", router)
	}()

	err := waitForServer(ctx, fmt.Sprintf("%s/ping", s.baseURL))
	s.Require().NoError(err)
}

func (s *AuthIntegrationSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(s.T().Context())
		s.Require().NoError(err)
	}
}

func (s *AuthIntegrationSuite) TestAuthenticate() {
	tests := []struct {
		name       string
		args       apphttp.AuthenticateArgs
		wantStatus int
	}{
		{
			name: "existing user",
			args: apphttp.AuthenticateArgs{
				Username: "first-username",
				Password: "first-username",
			},
			wantStatus: http.StatusOK,
		}, {
			name: "wrong password",
			args: apphttp.AuthenticateArgs{
				Username: "first-username",
				Password: "<wrong password>",
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	client := resty.New()
	defer client.Close()

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := client.R().
				SetBody(tt.args).
				Post(fmt.Sprintf("%s/api/login", s.baseURL))

			s.Require().NoError(err)
			var body []byte
			_, _ = resp.Body.Read(body)
			s.Equalf(tt.wantStatus, resp.StatusCode(), "BODY RESPONSE: %s", string(body))
		})
	}
}
