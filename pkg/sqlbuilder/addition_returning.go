package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

// ReturningAddition 表示 RETURNING 附加子句。
type ReturningAddition interface {
	Addition
}

// Returning 创建 RETURNING 附加子句。
func Returning(project sqlfrag.Fragment) ReturningAddition {
	return &returning{project: project}
}

type returning struct {
	project sqlfrag.Fragment
}

func (l *returning) AdditionType() AdditionType {
	return AdditionReturning
}

func (l *returning) IsNil() bool {
	return l == nil || sqlfrag.IsNil(l.project)
}

func (l *returning) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return sqlfrag.Pair("RETURNING ?", l.project).Frag(ContextWithToggles(ctx, Toggles{
		ToggleInProject: true,
	}))
}
