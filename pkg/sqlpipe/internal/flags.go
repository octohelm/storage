package internal

import (
	"context"

	contextx "github.com/octohelm/x/context"
)

var FlagsContext = contextx.New[Flags]()

type WithFlags interface {
	GetFlags(ctx context.Context) Flags
}

type Flags struct {
	OptWhereRequired *bool
	OptIncludesAll   *bool
	OptForReturning  *bool
}

func (f Flags) Merge(f2 Flags) Flags {
	if f2.OptIncludesAll != nil {
		f.OptIncludesAll = f2.OptIncludesAll
	}
	if f2.OptWhereRequired != nil {
		f.OptWhereRequired = f2.OptWhereRequired
	}
	if f2.OptForReturning != nil {
		f.OptForReturning = f2.OptForReturning
	}
	return f
}

func (f Flags) IncludesAll() bool {
	if f.OptIncludesAll == nil {
		return false
	}
	return *f.OptIncludesAll
}

func (f Flags) WhereRequired() bool {
	if f.OptWhereRequired == nil {
		return false
	}
	return *f.OptWhereRequired
}

func (f Flags) ForReturning() bool {
	if f.OptForReturning == nil {
		return false
	}
	return *f.OptForReturning
}

var _ WithFlags = Seed{}

type Seed struct {
	Flags
}

func (s Seed) GetFlags(ctx context.Context) Flags {
	if f, ok := FlagsContext.MayFrom(ctx); ok {
		return s.Flags.Merge(f)
	}
	return s.Flags
}
