package testutil

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"

	testingx "github.com/octohelm/x/testing"
)

func BeFragment(query string, args ...any) testingx.Matcher[sqlfrag.Fragment] {
	return &fragmentMatcher[sqlfrag.Fragment]{
		query: strings.TrimSpace(query),
		args:  args,
	}
}

type fragmentMatcher[A sqlfrag.Fragment] struct {
	query string
	args  []any
}

func (m *fragmentMatcher[A]) Negative() bool {
	return false
}

func (m *fragmentMatcher[A]) Name() string {
	return "Be Frag"
}

func (m *fragmentMatcher[A]) Match(actual A) bool {
	if sqlfrag.IsNil(actual) {
		return m.query == ""
	}
	q, args := sqlfrag.Collect(context.Background(), actual)
	if len(m.args) == 0 && len(args) == 0 {
		return m.query == q
	}

	return m.query == q && reflect.DeepEqual(m.args, args)
}

func (m *fragmentMatcher[A]) FormatActual(actual A) string {
	if sqlfrag.IsNil(actual) {
		return ""
	}
	q, args := sqlfrag.Collect(context.Background(), actual)
	return fmt.Sprintf("%s | %v", q, args)
}

func (m *fragmentMatcher[A]) FormatExpected() string {
	return fmt.Sprintf("%s | %v", m.query, m.args)
}
