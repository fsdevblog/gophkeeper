package pgrepo

import (
	"errors"
	"fmt"

	"github.com/fsdevblog/gophkeeper/internal/storage/repos"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
)

// Database error codes for PostgreSQL specific errors.
const (
	uniqueViolationCode = "23505"
	isolationCode       = "40001"
)

// convertErr transforms database errors into standardized repository layer errors.
// It adds formatted context message, business error type, and original error message.
//
// Features:
//   - Converts pgx.ErrNoRows to domain's ErrRecordNotFound
//   - Maps PostgreSQL unique constraint violations (code 23505) to domain's ErrDuplicateKey
//   - Maps PostgreSQL isolation errors (code 40001) to domain's ErrIsolation
//   - All other errors are wrapped as ErrUnknown with the original message
//
// Parameters:
//   - err: original error to convert
//   - format: message format string
//   - formatArgs: format arguments for the message
//
// Returns:
//   - nil if input error is nil
//   - wrapped and converted error with appropriate type and context
//
// The returned error format is: [repository/{context message}] {error type}: {original error}.
func convertErr(err error, format string, formatArgs ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, formatArgs...)

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("[repository/%s] %w", msg, repos.ErrRecordNotFound)
	}

	var pgErr *pgconn.PgError
	errType := repos.ErrUnknown

	if errors.As(err, &pgErr) {
		if isUniqueViolationErr(pgErr) {
			errType = repos.ErrDuplicateKey
		} else if isIsolationErr(pgErr) {
			errType = repos.ErrIsolation
		}
	}

	return fmt.Errorf("[repository/%s] %w: %w", msg, errType, err)
}

// isUniqueViolationErr checks if the given PostgreSQL error represents
// a unique constraint violation.
//
// Parameters:
//   - err: PostgreSQL error to check
//
// Returns:
//   - true if error code matches uniqueViolationCode (23505)
//   - false otherwise
func isUniqueViolationErr(err *pgconn.PgError) bool {
	return err.Code == uniqueViolationCode
}

// isIsolationErr checks if the given PostgreSQL error represents
// a transaction isolation error.
//
// Parameters:
//   - err: PostgreSQL error to check
//
// Returns:
//   - true if error code matches isolationCode (40001)
//   - false otherwise
func isIsolationErr(err *pgconn.PgError) bool {
	return err.Code == isolationCode
}
