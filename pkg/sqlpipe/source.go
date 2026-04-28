package sqlpipe

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

// Model 复用 sqlbuilder 的模型约束。
type Model = sqlbuilder.Model

// ModelNewer 为泛型数据源包装暴露类型化模型分配能力。
type ModelNewer[M Model] sqlbuilder.ModelNewer[M]

// Source 表示可组合的 SQL 产出阶段。
type Source[M Model] interface {
	Pipe(operators ...SourceOperator[M]) Source[M]

	sqlfrag.Fragment

	internal.StmtPatcher[M]
}

// SourceOperator 把一个 Source 转换为另一个 Source。
type SourceOperator[M Model] interface {
	internal.Operator[Source[M], Source[M]]

	OperatorType() OperatorType
}

// SourceOperatorFunc 用简单转换函数构造 SourceOperator。
func SourceOperatorFunc[M Model](typ OperatorType, next func(in Source[M]) Source[M]) SourceOperator[M] {
	return &operatorFunc[M]{typ: typ, next: next}
}

type operatorFunc[M Model] struct {
	typ  OperatorType
	next func(in Source[M]) Source[M]
}

func (fn *operatorFunc[M]) OperatorType() OperatorType {
	return fn.typ
}

func (fn *operatorFunc[M]) Next(src Source[M]) Source[M] {
	return fn.next(src)
}

// Pipe 按从左到右顺序依次应用操作符。
func Pipe[M Model](src Source[M], operators ...SourceOperator[M]) Source[M] {
	for _, o := range operators {
		src = o.Next(src)
	}
	return src
}
