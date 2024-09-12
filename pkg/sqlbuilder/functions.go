package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Count(fragments ...sqlfrag.Fragment) *Function {
	if len(fragments) == 0 {
		return Func("COUNT", sqlfrag.Pair("1"))
	}
	return Func("COUNT", fragments...)
}

func Avg(fragments ...sqlfrag.Fragment) *Function {
	return Func("AVG", fragments...)
}

func AnyValue(fragments ...sqlfrag.Fragment) *Function {
	return Func("ANY_VALUE", fragments...)
}

func Distinct(fragments ...sqlfrag.Fragment) *Function {
	return Func("DISTINCT", fragments...)
}

func Min(fragments ...sqlfrag.Fragment) *Function {
	return Func("MIN", fragments...)
}

func Max(fragments ...sqlfrag.Fragment) *Function {
	return Func("MAX", fragments...)
}

func First(fragments ...sqlfrag.Fragment) *Function {
	return Func("FIRST", fragments...)
}

func Last(fragments ...sqlfrag.Fragment) *Function {
	return Func("LAST", fragments...)
}

func Sum(fragments ...sqlfrag.Fragment) *Function {
	return Func("SUM", fragments...)
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
			for q, args := range sqlfrag.InlineBlock(sqlfrag.Const('*')).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
			return
		}

		for q, args := range sqlfrag.InlineBlock(sqlfrag.JoinValues(",", f.args...)).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
