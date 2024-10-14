package scanner

import (
	"context"
	"database/sql"

	"errors"

	"github.com/octohelm/storage/pkg/dberr"
)

func Scan(ctx context.Context, rows *sql.Rows, v interface{}) error {
	if rows == nil {
		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	si, err := ScanIteratorFor(v)
	if err != nil {
		return err
	}

	for rows.Next() {
		item := si.New()

		if scanErr := tryScan(ctx, rows, item); scanErr != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return scanErr
		}

		if err := si.Next(item); err != nil {
			return err
		}
	}

	if mustHasRecord, ok := si.(interface{ MustHasRecord() bool }); ok {
		if !mustHasRecord.MustHasRecord() {
			return dberr.New(dberr.ErrTypeNotFound, "record is not found")
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Make sure the query can be processed to completion with no errors.
	if err := rows.Close(); err != nil {
		return err
	}

	return nil
}

func tryScan(ctx context.Context, rows *sql.Rows, item any) error {
	done := make(chan error)
	go func() {
		defer close(done)

		if err := scanTo(ctx, rows, item); err != nil {
			done <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return <-done
	}
}
