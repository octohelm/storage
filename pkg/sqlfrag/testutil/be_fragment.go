package testutil

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"

	testingx "github.com/octohelm/x/testing"
)

func BeFragmentForQuery(query string, args ...any) testingx.Matcher[sqlfrag.Fragment] {
	return &fragmentMatcher[sqlfrag.Fragment]{
		query: strings.TrimSpace(query),
		args:  args,
	}
}

func BeFragment(query string, args ...any) testingx.Matcher[sqlfrag.Fragment] {
	return &fragmentMatcher[sqlfrag.Fragment]{
		query:     strings.TrimSpace(query),
		args:      args,
		checkArgs: true,
	}
}

type fragmentMatcher[A sqlfrag.Fragment] struct {
	query     string
	args      []any
	checkArgs bool

	queryNotEqual bool
	argsNotEqual  bool
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

	if m.query == q {
		return true
	}

	m.queryNotEqual = true

	if m.checkArgs {
		return reflect.DeepEqual(m.args, args)
	}

	m.argsNotEqual = true

	return false
}

func (m *fragmentMatcher[A]) FormatActual(actual A) string {
	if sqlfrag.IsNil(actual) {
		return ""
	}
	q, args := sqlfrag.Collect(context.Background(), actual)

	if m.queryNotEqual && m.argsNotEqual {
		return fmt.Sprintf("%s | %v", q, args)
	}

	if m.queryNotEqual {
		return q
	}

	return fmt.Sprintf("%v", args)
}

func (m *fragmentMatcher[A]) FormatExpected() string {
	if m.queryNotEqual && m.argsNotEqual {
		return fmt.Sprintf("%s | %v", m.query, m.args)
	}

	if m.queryNotEqual {
		return m.query
	}

	return fmt.Sprintf("%v", m.args)
}
