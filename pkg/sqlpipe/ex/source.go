package ex

import (
	"cmp"
	"context"
	"fmt"
	"iter"
	"slices"
	"sync"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlpipe"
	exiternal "github.com/octohelm/storage/pkg/sqlpipe/ex/internal"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
)

type SourceExecutor[M sqlpipe.Model] interface {
	sqlpipe.Source[M]

	// From to load data source other source for inserting
	From(source sqlpipe.Source[M]) SourceExecutor[M]
	// PipeE compose operators as SourceExecutor
	PipeE(operators ...sqlpipe.SourceOperator[M]) SourceExecutor[M]
	// Commit just execute
	Commit(ctx context.Context) error

	// Items execute and return Model or error
	Items(ctx context.Context) iter.Seq2[*M, error]
	// FindOne execute then scan as Model or error
	// notice this will return nil when not result found
	FindOne(ctx context.Context) (*M, error)
	// List execute then list as item list
	List(ctx context.Context) ([]*M, error)
	// ListTo execute then list to item adder
	ListTo(ctx context.Context, adder Adder[M]) error
	// CountTo execute as count
	CountTo(ctx context.Context, x *int64) error
}

type Adder[M sqlpipe.Model] interface {
	Add(m *M)
}

func FromSource[M sqlpipe.Model](src sqlpipe.Source[M]) SourceExecutor[M] {
	return &Executor[M]{
		src: src,
	}
}

type Executor[M sqlpipe.Model] struct {
	src       sqlpipe.Source[M]
	operators []sqlpipe.SourceOperator[M]

	once      sync.Once
	prepared  sqlpipe.Source[M]
	forCommit bool
}

func (e *Executor[M]) adapterOf(ctx context.Context, s session.Session) adapter.Adapter {
	if e.forCommit {
		return s.Adapter()
	}
	if session.InTx(ctx) {
		return s.Adapter()
	}
	return s.Adapter(session.ReadOnly())
}

// IsNil if true will omit in sql building
func (e *Executor[M]) IsNil() bool {
	return e.source().IsNil()
}

// Frag  for sql building
func (e *Executor[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return e.source().Frag(ctx)
}

// ApplyStmt for Builder
func (e *Executor[M]) ApplyStmt(ctx context.Context, s *internal.Builder[M]) *internal.Builder[M] {
	return e.source().ApplyStmt(ctx, s)
}

// Tx begin transaction if not in some transaction
func (e *Executor[M]) Tx(ctx context.Context, do func(ctx context.Context) error) error {
	return e.session(ctx).Tx(ctx, do)
}

// From to load data source other source for inserting
func (e *Executor[M]) From(src sqlpipe.Source[M]) SourceExecutor[M] {
	return FromSource(src)
}

// Pipe compose operators as Source
func (e *Executor[M]) Pipe(operators ...sqlpipe.SourceOperator[M]) sqlpipe.Source[M] {
	return e.source().Pipe(operators...)
}

// PipeE compose operators as SourceExecutor
func (e *Executor[M]) PipeE(operators ...sqlpipe.SourceOperator[M]) SourceExecutor[M] {
	if operators == nil {
		return e
	}

	return &Executor[M]{
		src:       e.src,
		operators: slices.Concat(e.operators, operators),
	}
}

func (e *Executor[M]) session(ctx context.Context) session.Session {
	m := new(M)
	s := session.For(ctx, m)
	if s == nil {
		panic(fmt.Errorf("invalid model %T", m))
	}
	return s
}

func (e *Executor[M]) source() sqlpipe.Source[M] {
	e.once.Do(func() {
		s := e.src
		if s == nil {
			s = sqlpipe.From[M]()
		}

		if len(e.operators) > 0 {
			operators := slices.SortedFunc(flatten(slices.Values(e.operators)), func(a sqlpipe.SourceOperator[M], b sqlpipe.SourceOperator[M]) int {
				return cmp.Compare(a.OperatorType(), b.OperatorType())
			})

			for _, op := range operators {
				if op.OperatorType() == sqlpipe.OperatorCommit {
					e.forCommit = true
					break
				}
			}

			s = s.Pipe(operators...)
		}

		e.prepared = s
	})

	return e.prepared
}

type collector[M sqlpipe.Model] struct {
	operators []sqlpipe.SourceOperator[M]
}

func (c *collector[M]) IsNil() bool {
	return false
}

func (c *collector[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return nil
}

func (c *collector[M]) ApplyStmt(ctx context.Context, s *internal.Builder[M]) *internal.Builder[M] {
	return nil
}

func (c *collector[M]) Pipe(operators ...sqlpipe.SourceOperator[M]) sqlpipe.Source[M] {
	if len(operators) > 0 {
		c.operators = append(c.operators, operators...)
	}
	return c
}

func flatten[M sqlpipe.Model](opSeq iter.Seq[sqlpipe.SourceOperator[M]]) iter.Seq[sqlpipe.SourceOperator[M]] {
	return func(yield func(sqlpipe.SourceOperator[M]) bool) {
		for op := range opSeq {
			if op.OperatorType() == sqlpipe.OperatorCompose {
				c := &collector[M]{}

				op.Next(c)

				for _, subOp := range c.operators {
					if !yield(subOp) {
						return
					}
				}

				continue
			}

			if !yield(op) {
				return
			}
		}
	}
}

// Commit execute sql
func (e *Executor[M]) Commit(ctx context.Context) error {
	// always for mutating
	e.forCommit = true
	_, err := e.adapterOf(ctx, e.session(ctx)).Exec(ctx, e.source())
	return err
}

// Items execute sql and returns Model or error as iter.Seq
func (e *Executor[M]) Items(ctx context.Context) iter.Seq2[*M, error] {
	s := e.session(ctx)

	ex := e.source().Pipe(
		sqlpipe.DefaultProject[M](internal.ColumnsByStruct(new(M))),
	)

	x := scanner.RecvFunc[M](func(ctx context.Context, recv func(v *M) error) error {
		rows, err := e.adapterOf(ctx, s).Query(internal.FlagContext.Inject(ctx, flags.ForReturning), ex)
		if err != nil {
			return err
		}
		return scanner.Scan(ctx, rows, scanner.Recv(recv))
	})

	return x.Items(ctx)
}

// List execute sql and return Model slice
func (e *Executor[M]) List(ctx context.Context) ([]*M, error) {
	list := make([]*M, 0)

	for item, err := range e.Items(ctx) {
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

// CountTo execute sql for count and marshal value into int64 ptr
func (e *Executor[M]) CountTo(ctx context.Context, x *int64) error {
	s := e.session(ctx)

	ex := e.source().Pipe(
		sqlpipe.Project[M](sqlbuilder.Count()),
		exiternal.ForCount[M](),
	)

	rows, err := e.adapterOf(ctx, s).Query(ctx, ex)
	if err != nil {
		return err
	}
	return scanner.Scan(ctx, rows, x)
}

// ListTo execute sql and return add Model into some Adder
func (e *Executor[M]) ListTo(ctx context.Context, listAdder Adder[M]) error {
	for item, err := range e.Items(ctx) {
		if err != nil {
			return err
		}
		listAdder.Add(item)
	}
	return nil
}

// FindOne execute then scan as Model or error
// notice this will return nil when not result found
func (e *Executor[M]) FindOne(ctx context.Context) (*M, error) {
	for item, err := range e.PipeE(sqlpipe.Limit[M](1)).Items(ctx) {
		if err != nil {
			return nil, err
		}
		return item, err
	}
	return nil, nil
}
