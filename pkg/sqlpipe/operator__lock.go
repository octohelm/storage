package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

type OnLockConflict uint8

const (
	OnLockConflictWait OnLockConflict = iota
	OnLockConflictNoWait
	OnLockConflictSkipLocked
)

type LockOption interface {
	SetOnLockConflict(onLockConflict OnLockConflict)
}

func SkipLocked() func(LockOption) {
	return func(o LockOption) {
		o.SetOnLockConflict(OnLockConflictSkipLocked)
	}
}

func NoWait() func(LockOption) {
	return func(o LockOption) {
		o.SetOnLockConflict(OnLockConflictNoWait)
	}
}

func ForUpdate[M Model](optionFns ...func(LockOption)) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorLock, func(src Source[M]) Source[M] {
		op := &lockedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			lockFor: lockForUpdate,
		}
		op.build(optionFns...)
		return op
	})
}

func ForNoKeyUpdate[M Model](optionFns ...func(LockOption)) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorLock, func(src Source[M]) Source[M] {
		op := &lockedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			lockFor: lockForNoKeyUpdate,
		}
		op.build(optionFns...)
		return op
	})
}

func ForShare[M Model](optionFns ...func(LockOption)) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorLock, func(src Source[M]) Source[M] {
		op := &lockedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			lockFor: lockForShare,
		}
		op.build(optionFns...)
		return op
	})
}

func ForKeyShare[M Model](optionFns ...func(LockOption)) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorLock, func(src Source[M]) Source[M] {
		op := &lockedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			lockFor: lockForKeyShare,
		}
		op.build(optionFns...)
		return op
	})
}

type lockedSource[M Model] struct {
	Embed[M]

	lockFor        lockFor
	onLockConflict OnLockConflict
}

type lockFor uint8

const (
	lockForUpdate lockFor = iota
	lockForNoKeyUpdate
	lockForShare
	lockForKeyShare
)

func (source *lockedSource[M]) SetOnLockConflict(onLockConflict OnLockConflict) {
	source.onLockConflict = onLockConflict
}

func (source *lockedSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, source)
}

func (source *lockedSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	s := session.For(ctx, new(M))
	if s != nil {
		// skip which db not support row lock
		switch s.Adapter().DriverName() {
		case "sqlite":
			return b
		}
	}

	a := sqlbuilder.AsAddition(sqlbuilder.AdditionLock, sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
		return func(yield func(string, []any) bool) {
			switch source.lockFor {
			case lockForUpdate:
				if !yield("FOR UPDATE", nil) {
					return
				}
			case lockForNoKeyUpdate:
				if !yield("FOR NO KEY UPDATE", nil) {
					return
				}
			case lockForShare:
				if !yield("FOR SHARE", nil) {
					return
				}
			case lockForKeyShare:
				if !yield("FOR KEY SHARE", nil) {
					return
				}
			}

			switch source.onLockConflict {
			case OnLockConflictNoWait:
				if !yield(" NO WAIT", nil) {
					return
				}
			case OnLockConflictSkipLocked:
				if !yield(" SKIP LOCKED", nil) {
					return
				}
			default:
			}
		}
	}))

	return source.Underlying.ApplyStmt(ctx, b.WithAdditions(a))
}

func (source *lockedSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](source, operators...)
}

func (source *lockedSource[M]) String() string {
	return internal.ToString(source)
}

func (source *lockedSource[M]) build(optionFns ...func(LockOption)) {
	for _, optFn := range optionFns {
		optFn(source)
	}
}
