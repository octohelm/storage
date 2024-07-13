package compose

import "github.com/octohelm/storage/pkg/sqlbuilder"

type From[M sqlbuilder.Model] struct {
}

func (f *From[M]) Model() *M {
	return new(M)
}
