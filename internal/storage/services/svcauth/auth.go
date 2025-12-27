package svcauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/storage/repos"

	"github.com/google/uuid"

	"github.com/fsdevblog/gophkeeper/internal/tokens"

	"github.com/fsdevblog/gophkeeper/internal/domain/models"
	repodto "github.com/fsdevblog/gophkeeper/internal/storage/repos/dto"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth/psswd"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
)

// JWTTokenExpiration defines the default JWT token lifetime (10 minutes).
const JWTTokenExpiration = 10 * time.Minute

// Options contains configuration settings for the authentication service.
type Options struct {
	// JWTTokenExpiration sets the expiration time for JWT tokens.
	JWTTokenExpiration time.Duration
}

// AuthService represents the authentication service.
type AuthService struct {
	// uow provides access to Unit of Work for repository operations.
	uow uow.UOW
	// psswdHasher provides password hashing functionality.
	psswdHasher PasswordHasher
	// jwtTokenSecret contains the secret key for JWT token generation.
	jwtTokenSecret []byte
	// tokenExpire defines the lifetime of generated tokens.
	tokenExpire time.Duration
}

// New creates a new instance of the authentication service.
//
// Parameters:
//   - u: Unit of Work interface for repository access
//   - jwtSecret: secret key for JWT token generation
//   - opts: optional functions for service configuration
//
// Returns:
//   - pointer to AuthService
//   - error if initialization fails
func New(u uow.UOW, jwtSecret []byte, opts ...func(*Options)) *AuthService {
	options := Options{
		JWTTokenExpiration: JWTTokenExpiration,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &AuthService{
		uow:            u,
		jwtTokenSecret: jwtSecret,
		psswdHasher:    new(psswd.PasswordHash),
	}
}

type DeviceArgs struct {
	DeviceType      models.DeviceType
	DeviceHash      uuid.UUID
	Platform        string
	PlatformVersion string
	AppVersion      string
}

// AuthenticateArgs contains arguments for the Authenticate method.
//
// Fields:
//   - Username: user's username
//   - Password: user's password
//   - Device: device information of the user.
type AuthenticateArgs struct {
	Username string
	Password string
	Device   DeviceArgs
}

// Authenticate verifies user credentials and generates a JWT token.
//
// Parameters:
//   - ctx: execution context
//   - args: see AuthenticateArgs for details
//
// Returns:
//   - JWT token as string
//   - pointer to user model
//   - error in case of authentication failure (ErrInvalidCredentials) or other issues
func (a *AuthService) Authenticate(ctx context.Context, args AuthenticateArgs) (string, *models.User, error) {
	var user *models.User
	var token string

	err := a.uow.Do(ctx, func(doCtx context.Context, tx uow.TX) error {
		var errUser, errToken error
		userRepo, errUserRepo := uow.GetAs[UserRepository](tx, uow.RepoName(repodto.UserRepoName))
		if errUserRepo != nil {
			return errUserRepo //nolint:wrapcheck
		}
		user, errUser = userRepo.FindByUsername(doCtx, args.Username)
		if errUser != nil {
			if errors.Is(errUser, repos.ErrRecordNotFound) {
				return ErrInvalidCredentials
			}
			return errUser //nolint:wrapcheck
		}
		if !a.psswdHasher.ComparePassword(args.Password, user.EncryptedPassword) {
			return ErrInvalidCredentials
		}

		if errDevice := a.createDevice(doCtx, tx, user.ID, args.Device); errDevice != nil {
			return errDevice //nolint:wrapcheck
		}

		token, errToken = a.genToken(user)
		if errToken != nil {
			return errToken //nolint:wrapcheck
		}
		return nil
	})

	if err != nil {
		return "", nil, fmt.Errorf("authenticate: %w", err)
	}

	return token, user, nil
}

// RegisterArgs contains arguments for the Register method.
//
// Fields:
//   - Username: username for the new user.
//   - Password: password for the new user.
//   - Device: device information of the new user.
type RegisterArgs struct {
	Username string
	Password string
	Device   DeviceArgs
}

// Register creates a new user account and generates a JWT token.
//
// Parameters:
//   - ctx: execution context
//   - args: registration arguments containing username and password
//
// Returns:
//   - JWT token as string
//   - pointer to the newly created user model
//   - error in case of registration failure (ErrUserAlreadyRegistered) or other issues
func (a *AuthService) Register(ctx context.Context, args RegisterArgs) (string, *models.User, error) {
	var user *models.User
	var token string
	err := a.uow.Do(ctx, func(doCtx context.Context, tx uow.TX) error {
		var errCreateUser, errToken error

		encryptedPassword, errPass := a.psswdHasher.HashPassword(args.Password)
		if errPass != nil {
			return errPass //nolint:wrapcheck
		}

		user, errCreateUser = a.createUser(doCtx, tx, repodto.CreateUserArgs{
			Username:          args.Username,
			EncryptedPassword: encryptedPassword,
		})
		if errCreateUser != nil {
			return errCreateUser //nolint:wrapcheck
		}

		if errCreateDevice := a.createDevice(doCtx, tx, user.ID, args.Device); errCreateDevice != nil {
			return errCreateDevice //nolint:wrapcheck
		}

		token, errToken = a.genToken(user)
		if errToken != nil {
			return errToken //nolint:wrapcheck
		}

		return nil
	})

	if err != nil {
		return "", nil, fmt.Errorf("register: %w", err)
	}
	return token, user, nil
}

func (a *AuthService) createDevice(
	ctx context.Context,
	tx uow.TX,
	userID uuid.UUID,
	args DeviceArgs,
) error {
	deviceRepo, errRepo := uow.GetAs[DeviceRepository](tx, uow.RepoName(repodto.DeviceRepoName))
	if errRepo != nil {
		return errRepo //nolint:wrapcheck
	}

	err := deviceRepo.Create(ctx, userID, repodto.CreateDeviceArgs{
		DeviceType:      args.DeviceType,
		DeviceHash:      args.DeviceHash,
		Platform:        args.Platform,
		PlatformVersion: args.PlatformVersion,
		AppVersion:      args.AppVersion,
		StateVersion:    uuid.Nil,
	})
	if err != nil {
		if errors.Is(err, repos.ErrDuplicateKey) {
			return nil
		}
		return err //nolint:wrapcheck
	}
	return nil
}

func (a *AuthService) createUser(ctx context.Context, tx uow.TX, args repodto.CreateUserArgs) (*models.User, error) {
	userRepo, errUserRepo := uow.GetAs[UserRepository](tx, uow.RepoName(repodto.UserRepoName))
	if errUserRepo != nil {
		return nil, errUserRepo // nolint:wrapcheck
	}
	user, errCreateUser := userRepo.Create(ctx, args)
	if errCreateUser != nil {
		if errors.Is(errCreateUser, repos.ErrDuplicateKey) {
			return nil, ErrUserAlreadyRegistered
		}
		return nil, errCreateUser //nolint:wrapcheck
	}
	return user, nil
}

// genToken generates a JWT token for the given user.
//
// Parameters:
//   - user: pointer to the user model requiring token generation
//
// Returns:
//   - string: generated JWT token
//   - error: if user is nil, has empty ID, or token generation fails
func (a *AuthService) genToken(user *models.User) (string, error) {
	if user == nil {
		return "", errors.New("generate token: user is nil")
	}

	if user.ID.String() == "" {
		return "", errors.New("generate token: user ID is empty")
	}
	token, err := tokens.GenerateUserJWT(user.ID, a.tokenExpire, a.jwtTokenSecret)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}
