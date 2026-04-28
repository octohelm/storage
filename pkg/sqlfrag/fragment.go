package sqlfrag

import (
	"bytes"
	"context"
	"iter"
	"slices"
	"strings"
)

// IsNil 判断片段是否为空。
func IsNil(e Fragment) bool {
	return e == nil || e.IsNil()
}

// Fragment 表示可展开为 SQL 文本与参数的片段。
type Fragment interface {
	IsNil() bool
	Frag(ctx context.Context) iter.Seq2[string, []any]
}

// Collect 收集片段生成的 SQL 文本与参数。
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
