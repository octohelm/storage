package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

// Count 创建 COUNT 聚合函数。
func Count(fragments ...sqlfrag.Fragment) *Function {
	if len(fragments) == 0 {
		return Func("COUNT", sqlfrag.Pair("1"))
	}
	return Func("COUNT", fragments...)
}

// Avg 创建 AVG 聚合函数。
func Avg(fragments ...sqlfrag.Fragment) *Function {
	return Func("AVG", fragments...)
}

// AnyValue 创建 ANY_VALUE 聚合函数。
func AnyValue(fragments ...sqlfrag.Fragment) *Function {
	return Func("ANY_VALUE", fragments...)
}

// Distinct 创建 DISTINCT 包装函数。
func Distinct(fragments ...sqlfrag.Fragment) *Function {
	return Func("DISTINCT", fragments...)
}

// Min 创建 MIN 聚合函数。
func Min(fragments ...sqlfrag.Fragment) *Function {
	return Func("MIN", fragments...)
}

// Max 创建 MAX 聚合函数。
func Max(fragments ...sqlfrag.Fragment) *Function {
	return Func("MAX", fragments...)
}

// First 创建 FIRST 聚合函数。
func First(fragments ...sqlfrag.Fragment) *Function {
	return Func("FIRST", fragments...)
}

// Last 创建 LAST 聚合函数。
func Last(fragments ...sqlfrag.Fragment) *Function {
	return Func("LAST", fragments...)
}

// Sum 创建 SUM 聚合函数。
func Sum(fragments ...sqlfrag.Fragment) *Function {
	return Func("SUM", fragments...)
}

// Func 按名称与参数创建通用函数调用。
func Func(name string, args ...sqlfrag.Fragment) *Function {
	if name == "" {
		return nil
	}
	return &Function{
		name: name,
		args: args,
	}
}

// Function 表示一个 SQL 函数调用。
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
