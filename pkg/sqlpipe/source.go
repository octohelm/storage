package sqlpipe

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

type Model = sqlbuilder.Model

type ModelNewer[M Model] sqlbuilder.ModelNewer[M]

type Source[M Model] interface {
	Pipe(operators ...SourceOperator[M]) Source[M]

	sqlfrag.Fragment

	internal.StmtPatcher[M]
}

type SourceOperator[M Model] interface {
	internal.Operator[Source[M], Source[M]]

	OperatorType() OperatorType
}

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

func Pipe[M Model, Operator interface{ SourceOperator[M] }](src Source[M], operators ...Operator) Source[M] {
	for _, o := range operators {
		src = o.Next(src)
	}
	return src
}
