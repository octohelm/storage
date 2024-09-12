package sqlfrag

import (
	"bytes"
	"context"
	"iter"
	"slices"
	"strings"
)

func IsNil(e Fragment) bool {
	return e == nil || e.IsNil()
}

type Fragment interface {
	IsNil() bool
	Frag(ctx context.Context) iter.Seq2[string, []any]
}

func Collect(ctx context.Context, f Fragment) (string, []any) {
	if f.IsNil() {
		return "", nil
	}

	b := bytes.NewBuffer(nil)
	args := make([]any, 0)

	for query, queryArgs := range f.Frag(ctx) {
		if len(query) > 0 {
			if query[0] == '\r' {
				b.WriteRune('\n')
				b.WriteString(query[1:])
			} else {
				b.WriteString(query)
			}
		}

		if len(queryArgs) > 0 {
			args = slices.Concat(args, queryArgs)
		}
	}

	q := b.String()

	return strings.TrimPrefix(q, "\n"), args
}
