package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	slicesx "github.com/octohelm/x/slices"
)

func DistinctOn[M Model](cols ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorProject, func(src Source[M]) Source[M] {
		if len(cols) == 0 {
			return src
		}
		return &distinctOn[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			on: slicesx.Map(cols, func(col modelscoped.Column[M]) sqlfrag.Fragment {
				return col
			}),
		}
	})
}

type distinctOn[M Model] struct {
	Embed[M]

	on []sqlfrag.Fragment
}

func (s *distinctOn[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *distinctOn[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	x := s.Underlying.ApplyStmt(ctx, b)

	if len(x.Orders) > 0 {
		m := *new(M)

		base := x.WithDistinctOn(s.on...)
		base.Pager = nil
		base.DefaultProjects = nil

		stmt := base.WithProjects(sqlfrag.Pair("?.*", sqlfrag.Const(m.TableName()))).BuildStmt(ctx)

		w := &internal.Builder[M]{}

		return w.WithFlag(s.GetFlag(ctx)).
			WithTableJoins(x.TableJoins...).
			WithOrders(x.Orders...).
			WithPager(x.Pager).
			WithDefaultProjects(x.DefaultProjects...).
			WithProjects(x.Projects...).
			WithSource(
				sqlfrag.Pair("? AS ?",
					sqlfrag.Block(stmt),
					sqlfrag.Const(m.TableName()),
				),
			)
	}

	return x.WithDistinctOn(s.on...)
}

func (s *distinctOn[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *distinctOn[M]) String() string {
	return internal.ToString(s)
}
