package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Count(sqlExprs ...sqlfrag.Fragment) *Function {
	if len(sqlExprs) == 0 {
		return Func("COUNT", sqlfrag.Pair("1"))
	}
	return Func("COUNT", sqlExprs...)
}

func Avg(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("AVG", sqlExprs...)
}

func Distinct(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("DISTINCT", sqlExprs...)
}

func Min(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("MIN", sqlExprs...)
}

func Max(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("MAX", sqlExprs...)
}

func First(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("FIRST", sqlExprs...)
}

func Last(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("LAST", sqlExprs...)
}

func Sum(sqlExprs ...sqlfrag.Fragment) *Function {
	return Func("SUM", sqlExprs...)
}

func Func(name string, args ...sqlfrag.Fragment) *Function {
	if name == "" {
		return nil
	}
	return &Function{
		name: name,
		args: args,
	}
}

type Function struct {
	name string
	args []sqlfrag.Fragment
}

func (f *Function) IsNil() bool {
	return f == nil || f.name == ""
}

func (f *Function) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield(f.name, nil) {
			return
		}

		if len(f.args) == 0 {
			for q, args := range sqlfrag.Group(sqlfrag.Const('*')).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
			return
		}

		for q, args := range sqlfrag.Group(sqlfrag.JoinValues(",", f.args...)).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
