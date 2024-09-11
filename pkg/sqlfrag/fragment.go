package sqlfrag

import (
	"bytes"
	"context"
	"iter"
	"slices"
)

func IsNil(e Fragment) bool {
	return e == nil || e.IsNil()
}

type Fragment interface {
	IsNil() bool
	Frag(ctx context.Context) iter.Seq2[string, []any]
}

func All(ctx context.Context, f Fragment) (string, []any) {
	if f.IsNil() {
		return "", nil
	}

	b := bytes.NewBuffer(nil)
	args := make([]any, 0)

	for query, queryArgs := range f.Frag(ctx) {
		b.WriteString(query)

		if len(queryArgs) > 0 {
			args = slices.Concat(args, queryArgs)
		}
	}

	return b.String(), args
}
