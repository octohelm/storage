package sqlpipe

import (
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
)

func Columns[M Model](cols ...modelscoped.Column[M]) iter.Seq[modelscoped.Column[M]] {
	return slices.Values(cols)
}
