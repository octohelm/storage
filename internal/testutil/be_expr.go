package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func ShouldBeExpr(t testing.TB, sqlExpr sqlbuilder.SqlExpr, query string, args ...interface{}) {
	t.Helper()

	if sqlbuilder.IsNilExpr(sqlExpr) {
		Expect(t, "", Be(strings.TrimSpace(query)))
		return
	}

	expr := sqlExpr.Ex(context.Background())

	Expect(t, expr.Query(), Be(strings.TrimSpace(query)))
	Expect(t, expr.Args(), Equal(args))
}
