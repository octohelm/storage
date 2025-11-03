package sqlpipe

import (
	"context"
	"iter"

	slicesx "github.com/octohelm/x/slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func Returning[M Model](cols ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorProject, func(src Source[M]) Source[M] {
		if len(cols) == 0 {
			return &projectedSource[M]{
				Embed: Embed[M]{
					Underlying: src,
				},
				projects: []sqlfrag.Fragment{
					sqlfrag.Const("*"),
				},
			}
		}

		return &projectedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			projects: slicesx.Map(cols, func(col modelscoped.Column[M]) sqlfrag.Fragment {
				return col
			}),
		}
	})
}

func Select[M Model](cols ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorProject, func(src Source[M]) Source[M] {
		if len(cols) == 0 {
			return src
		}
		return &projectedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			projects: slicesx.Map(cols, func(col modelscoped.Column[M]) sqlfrag.Fragment {
				return col
			}),
		}
	})
}

func CastSelect[M Model, U Model](cols ...modelscoped.Column[U]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorProject, func(src Source[M]) Source[M] {
		if len(cols) == 0 {
			return src
		}
		return &projectedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			projects: slicesx.Map(cols, func(col modelscoped.Column[U]) sqlfrag.Fragment {
				return col
			}),
		}
	})
}

func Project[M Model](projects ...sqlfrag.Fragment) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorProject, func(src Source[M]) Source[M] {
		return &projectedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			projects: projects,
		}
	})
}

func DefaultProject[M Model](projects ...sqlfrag.Fragment) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorSetting, func(src Source[M]) Source[M] {
		return &projectedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			projects:  projects,
			asDefault: true,
		}
	})
}

type projectedSource[M Model] struct {
	Embed[M]

	projects  []sqlfrag.Fragment
	asDefault bool
}

func (s *projectedSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *projectedSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	if s.asDefault {
		return s.Underlying.ApplyStmt(ctx, b.WithDefaultProjects(s.projects...))
	}
	return s.Underlying.ApplyStmt(ctx, b.WithProjects(s.projects...))
}

func (s *projectedSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *projectedSource[M]) String() string {
	return internal.ToString(s)
}
