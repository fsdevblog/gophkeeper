package services

import (
	"fmt"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/repos/pgrepo"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/config"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcentry"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Collection struct {
	config       *config.Config
	AuthService  *svcauth.AuthService
	EntryService *svcentry.EntryService
}

func NewCollection(config *config.Config, conn *pgxpool.Pool) (*Collection, error) {
	c := &Collection{
		config: config,
	}
	if err := c.initServices(conn); err != nil {
		return nil, fmt.Errorf("init service collection: %w", err)
	}
	return c, nil
}

func (c *Collection) initServices(conn *pgxpool.Pool) error {
	unitOfWork, errUOW := c.initUOW(conn)
	if errUOW != nil {
		return fmt.Errorf("init services: %w", errUOW)
	}
	c.AuthService = svcauth.New(unitOfWork, []byte(c.config.JWTSecret), func(opt *svcauth.Options) {
		opt.JWTTokenExpiration = time.Duration(c.config.JWTExpireInSeconds) * time.Second
	})
	c.EntryService = svcentry.New(unitOfWork)
	return nil
}

func (c *Collection) initUOW(conn *pgxpool.Pool) (*uow.UnitOfWork, error) {
	unitOfWork := uow.New(conn)
	var repos = make(map[uow.RepoName]uow.RepositoryFactory, 2)
	{
		repos[uow.RepoName(dto.UserRepoName)] = func(dbtx uow.DBTX) uow.Repository {
			return pgrepo.NewUserRepo(dbtx)
		}
		repos[uow.RepoName(dto.DeviceRepoName)] = func(dbtx uow.DBTX) uow.Repository {
			return pgrepo.NewDeviceRepo(dbtx)
		}
	}

	if err := unitOfWork.MassRegister(repos); err != nil {
		return nil, fmt.Errorf("register repositories: %w", err)
	}
	return unitOfWork, nil
}
