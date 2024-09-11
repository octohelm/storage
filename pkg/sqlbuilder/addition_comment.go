package sqlbuilder

import (
	"context"
	"iter"
)

func Comment(c string) Addition {
	return &comment{text: []byte(c)}
}

type comment struct {
	text []byte
}

func (comment) AdditionType() AdditionType {
	return AdditionComment
}

func (c *comment) IsNil() bool {
	return c == nil || len(c.text) == 0
}

func (c *comment) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("/* "+string(c.text)+" */", nil) {
			return
		}
	}
}
