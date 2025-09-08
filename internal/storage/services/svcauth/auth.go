package svcauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/domain"
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
	// userRepo represents the repository for user operations.
	userRepo UserRepository
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
func New(u uow.UOW, jwtSecret []byte, opts ...func(*Options)) (*AuthService, error) {
	options := Options{
		JWTTokenExpiration: JWTTokenExpiration,
	}

	for _, opt := range opts {
		opt(&options)
	}

	userRepo, errUserRepo := uow.GetRepositoryAs[UserRepository](u, uow.RepoName(repodto.UserRepoName))
	if errUserRepo != nil {
		return nil, fmt.Errorf("init auth service: %w", errUserRepo)
	}
	return &AuthService{
		uow:            u,
		userRepo:       userRepo,
		jwtTokenSecret: jwtSecret,
		psswdHasher:    new(psswd.PasswordHash),
	}, nil
}

// AuthenticateArgs contains arguments for the Authenticate method.
//
// Fields:
//   - Username: user's username
//   - Password: <PASSWORD>
type AuthenticateArgs struct {
	Username string
	Password string
}

// Authenticate verifies user credentials and generates a JWT token.
//
// Parameters:
//   - ctx: execution context
//   - username: user's username
//   - password: user's password
//
// Returns:
//   - JWT token as string
//   - pointer to user model
//   - error in case of authentication failure (ErrInvalidPassword) or other issues
func (a *AuthService) Authenticate(ctx context.Context, args AuthenticateArgs) (string, *models.User, error) {
	user, errUser := a.userRepo.FindByUsername(ctx, args.Username)
	if errUser != nil {
		return "", nil, fmt.Errorf("authenticate: %w", errUser)
	}
	if !a.psswdHasher.ComparePassword(args.Password, user.EncryptedPassword) {
		return "", nil, fmt.Errorf("authenticate: %w", ErrInvalidPassword)
	}

	token, errToken := a.genToken(user)
	if errToken != nil {
		return "", nil, fmt.Errorf("authenticate: %w", errToken)
	}
	return token, user, nil
}

// RegisterArgs contains arguments for the Register method.
//
// Fields:
//   - Username: desired username for the new user.
//   - Password: desired password for the new user (sensitive information).
type RegisterArgs struct {
	Username string
	Password string
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
	user, errCreate := a.userRepo.CreateUser(ctx, repodto.CreateUserArgs{
		Username: args.Username,
		Password: args.Password,
	})
	if errCreate != nil {
		if errors.Is(errCreate, domain.ErrDuplicateKey) {
			return "", nil, fmt.Errorf("register: %w", ErrUserAlreadyRegistered)
		}
		return "", nil, fmt.Errorf("register: %w", errCreate)
	}

	token, errToken := a.genToken(user)
	if errToken != nil {
		return "", nil, fmt.Errorf("register: %w", errToken)
	}
	return token, user, nil
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
