package target

import (
	"context"

	"github.com/octohelm/storage/deprecated/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Table[T sqlbuilder.Model](ctx context.Context) sqlbuilder.Table {
	m := new(T)
	s := dal.SessionFor(ctx, m)
	return &modelTable[T]{
		Table: s.T(m),
	}
}

type modelTable[T sqlbuilder.Model] struct {
	sqlbuilder.Table
}

func (m *modelTable[T]) New() sqlbuilder.Model {
	return *new(T)
}
