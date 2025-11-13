package testutil

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	testingx "github.com/octohelm/x/testing"
	"github.com/octohelm/x/testing/snapshot"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func BeFragmentForQuery(query string, args ...any) testingx.Matcher[sqlfrag.Fragment] {
	return &fragmentMatcher{
		query: strings.TrimSpace(query),
		args:  args,
	}
}

func BeFragment(query string, args ...any) testingx.Matcher[sqlfrag.Fragment] {
	return &fragmentMatcher{
		query:     strings.TrimSpace(query),
		args:      args,
		checkArgs: true,
	}
}

type fragmentMatcher struct {
	query     string
	args      []any
	checkArgs bool

	queryNotEqual bool
	argsNotEqual  bool
}

func (m *fragmentMatcher) Negative() bool {
	return false
}

func (m *fragmentMatcher) Action() string {
	return "Be Frag"
}

func (m *fragmentMatcher) Match(actual sqlfrag.Fragment) bool {
	if sqlfrag.IsNil(actual) {
		return m.query == ""
	}

	q, args := sqlfrag.Collect(context.Background(), actual)
	if len(m.args) == 0 && len(args) == 0 {
		if m.query == q {
			return true
		}
		m.queryNotEqual = true
		return false
	}

	if m.query == q {
		return true
	}

	m.queryNotEqual = true

	if m.checkArgs {
		if reflect.DeepEqual(m.args, args) {
			return true
		}
	}

	m.argsNotEqual = true

	return false
}

func (m *fragmentMatcher) NormalizeActual(actual sqlfrag.Fragment) any {
	if sqlfrag.IsNil(actual) {
		return ""
	}

	q, args := sqlfrag.Collect(context.Background(), actual)

	if m.queryNotEqual && m.argsNotEqual {
		return slices.Concat(
			snapshot.LinesFromBytes([]byte(q)),
			linesFromArgs(args),
		)
	}

	if m.queryNotEqual {
		return snapshot.LinesFromBytes([]byte(q))
	}

	return linesFromArgs(args)
}

var _ testingx.MatcherWithNormalizedExpected = &fragmentMatcher{}

func (m *fragmentMatcher) NormalizedExpected() any {
	if m.queryNotEqual && m.argsNotEqual {
		return slices.Concat(
			snapshot.LinesFromBytes([]byte(m.query)),
			linesFromArgs(m.args),
		)
	}

	if m.queryNotEqual {
		return snapshot.LinesFromBytes([]byte(m.query))
	}

	return linesFromArgs(m.args)
}

func linesFromArgs(args []any) snapshot.Lines {
	lines := make(snapshot.Lines, 0, len(args))
	for _, arg := range args {
		lines = append(lines, fmt.Sprintf("%v", arg))
	}
	return lines
}
