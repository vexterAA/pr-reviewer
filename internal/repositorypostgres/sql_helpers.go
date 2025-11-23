package repositorypostgres

import (
	"context"
	"database/sql"
	"errors"
)

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var errNotFound = sql.ErrNoRows

func closeRows(rows *sql.Rows) {
	if rows != nil {
		_ = rows.Close()
	}
}

func wrapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errNotFound
	}
	return err
}
