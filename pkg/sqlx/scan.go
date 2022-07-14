package sqlx

import (
	"context"
	"database/sql"

	"github.com/octohelm/sqlx/internal/scanner"
)

type ScanIterator = scanner.ScanIterator

func Scan(rows *sql.Rows, v interface{}) error {
	if err := scanner.Scan(context.Background(), rows, v); err != nil {
		return err
	}
	return nil
}
