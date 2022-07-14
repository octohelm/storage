package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/octohelm/sqlx/pkg/builder"
)

func ShouldBeExpr(t testing.TB, sqlExpr builder.SqlExpr, query string, args ...interface{}) {
	t.Helper()

	if builder.IsNilExpr(sqlExpr) {
		Expect(t, "", Be(strings.TrimSpace(query)))
		return
	}

	expr := sqlExpr.Ex(context.Background())

	Expect(t, expr.Query(), Be(strings.TrimSpace(query)))
	Expect(t, expr.Args(), Equal(args))
}
