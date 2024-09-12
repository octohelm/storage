package sqlpipe

import (
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"iter"
	"slices"
)

func Columns[M Model](cols ...modelscoped.Column[M]) iter.Seq[modelscoped.Column[M]] {
	return slices.Values(cols)
}
