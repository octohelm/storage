package migration

import (
	"context"
	"io"

	"github.com/octohelm/sqlx/pkg/sqlx"
	contextx "github.com/octohelm/x/context"
)

type contextKeyOutput struct{}

func OutputFromContext(ctx context.Context) io.Writer {
	if opts, ok := ctx.Value(contextKeyOutput{}).(io.Writer); ok {
		if opts != nil {
			return opts
		}
	}
	return nil
}

func Migrate(db sqlx.DBExecutor, output io.Writer) error {
	ctx := contextx.WithValue(db.Context(), contextKeyOutput{}, output)

	if err := db.(sqlx.Migrator).Migrate(ctx, db); err != nil {
		return err
	}
	return nil
}

func MustMigrate(db sqlx.DBExecutor, w io.Writer) {
	if err := Migrate(db, w); err != nil {
		panic(err)
	}
}
