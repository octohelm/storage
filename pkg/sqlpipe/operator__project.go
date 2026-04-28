package sqlpipe

import (
	"context"
	"iter"

	slicesx "github.com/octohelm/x/slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

// Returning 为数据源追加 RETURNING 投影。
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

// Select 为数据源指定显式投影列。
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

// CastSelect 为不同模型类型的列集合指定投影。
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

// Project 直接用片段列表指定投影。
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

// DefaultProject 设置默认投影。
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
