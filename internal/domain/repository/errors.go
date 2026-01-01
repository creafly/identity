package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/creafly/logger"
	"github.com/lib/pq"
)

var (
	ErrNotFound            = errors.New("record not found")
	ErrDuplicateEntry      = errors.New("record already exists")
	ErrForeignKeyViolation = errors.New("referenced record does not exist")
	ErrDatabaseOperation   = errors.New("database operation failed")
	ErrInvalidData         = errors.New("invalid data")
)

func wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	logger.Debug().
		Str("operation", operation).
		Str("errorType", getErrorType(err)).
		Msg("Database operation error")

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return ErrDuplicateEntry
		case "23503":
			return ErrForeignKeyViolation
		case "23502":
			return ErrInvalidData
		case "23514":
			return ErrInvalidData
		case "22001":
			return ErrInvalidData
		case "22P02":
			return ErrInvalidData
		}
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique constraint") {
		return ErrDuplicateEntry
	}
	if strings.Contains(errMsg, "foreign key") {
		return ErrForeignKeyViolation
	}
	if strings.Contains(errMsg, "no rows") {
		return ErrNotFound
	}

	return ErrDatabaseOperation
}

func getErrorType(err error) string {
	if err == nil {
		return "nil"
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return "pq:" + string(pqErr.Code)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return "sql:no_rows"
	}

	return "unknown"
}
