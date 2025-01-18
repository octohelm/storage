package sqlpipe

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	slicesx "github.com/octohelm/x/slices"
)

func AggregateGroupBy[S Model, O Model](src Source[S], by iter.Seq[modelscoped.Column[S]], cols ...modelscoped.Column[O]) Source[O] {
	return &aggregatedSource[S, O]{
		Embed: Embed[S]{
			Underlying: src,
		},
		groupBy: sqlfrag.NonNil(by),
		projects: slicesx.Map(cols, func(col modelscoped.Column[O]) sqlfrag.Fragment {
			return col
		}),
		projectsForDownstream: slicesx.Map(cols, func(col modelscoped.Column[O]) sqlfrag.Fragment {
			return col.ComputedBy(nil)
		}),
	}
}

func Aggregate[S Model, O Model](src Source[S], cols ...modelscoped.Column[O]) Source[O] {
	return &aggregatedSource[S, O]{
		Embed: Embed[S]{
			Underlying: src,
		},
		projects: slicesx.Map(cols, func(col modelscoped.Column[O]) sqlfrag.Fragment {
			return col
		}),
		projectsForDownstream: slicesx.Map(cols, func(col modelscoped.Column[O]) sqlfrag.Fragment {
			return col.ComputedBy(nil)
		}),
	}
}

type aggregatedSource[S Model, O Model] struct {
	Embed[S]
	projects              []sqlfrag.Fragment
	projectsForDownstream []sqlfrag.Fragment
	groupBy               iter.Seq[sqlfrag.Fragment]
}

func (s *aggregatedSource[S, O]) IsNil() bool {
	return s.Underlying == nil
}

func (s *aggregatedSource[S, O]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *aggregatedSource[S, O]) ApplyStmt(ctx context.Context, b *internal.Builder[O]) *internal.Builder[O] {
	b = b.WithFlag(s.GetFlag(ctx))

	b = b.WithProjects(s.projects...)

	if s.groupBy != nil {
		b = b.WithAdditions(sqlbuilder.GroupBy(slices.Collect(s.groupBy)...))
	}

	switch src := s.Underlying.(type) {
	case *sourceFrom[S]:
		return b.WithSource(sqlfrag.Const((*new(S)).TableName()))
	default:
		stmt := internal.BuildStmt(ctx, src)

		return b.WithSource(
			sqlfrag.Pair("? AS ?",
				sqlfrag.Block(stmt),
				sqlfrag.Const((*new(S)).TableName()),
			),
		)
	}
}

func (s *aggregatedSource[S, O]) Pipe(operators ...SourceOperator[O]) Source[O] {
	return Pipe[O](As[O](s, (*new(O)).TableName()), append(operators, DefaultProject[O](s.projectsForDownstream...))...)
}

func (s *aggregatedSource[S, O]) String() string {
	return internal.ToString(s)
}
