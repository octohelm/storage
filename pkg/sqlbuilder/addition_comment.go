package sqlbuilder

import (
	"context"
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

func (c *comment) Ex(ctx context.Context) *Ex {
	e := ExactlyExpr("")
	e.WhiteComments(c.text)
	return e.Ex(ctx)
}
