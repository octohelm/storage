package querierpatcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func ApplyTo[M sqlbuilder.Model](q dal.Querier, patchers ...Typed[M]) dal.Querier {
	return q.Apply(CastSliceAsAnySlice(patchers...)...)
}
