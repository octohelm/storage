package sqlpipe

import (
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
)

// Columns 把列切片包装为列序列。
func Columns[M Model](cols ...modelscoped.Column[M]) iter.Seq[modelscoped.Column[M]] {
	return slices.Values(cols)
}
