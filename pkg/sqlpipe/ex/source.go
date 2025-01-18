package ex

import (
	"cmp"
	"context"
	"fmt"
	"iter"
	"slices"

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

	From(source sqlpipe.Source[M]) SourceExecutor[M]

	PipeE(operators ...sqlpipe.SourceOperator[M]) SourceExecutor[M]

	// just execute
	Commit(ctx context.Context) error

	// execute then iter item
	Item(ctx context.Context) iter.Seq2[*M, error]
	// execute then range
	Range(ctx context.Context, fn func(*M)) error

	// execute then scan as one item
	FindOne(ctx context.Context) (*M, error)
	// execute then list as item list
	List(ctx context.Context) ([]*M, error)
	// execute then list to item adder
	ListTo(ctx context.Context, adder Adder[M]) error
	// execute as count
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
	src sqlpipe.Source[M]

	operators []sqlpipe.SourceOperator[M]
}

func (e *Executor[M]) Tx(ctx context.Context, do func(ctx context.Context) error) error {
	return e.session(ctx).Tx(ctx, do)
}

func (e *Executor[M]) From(src sqlpipe.Source[M]) SourceExecutor[M] {
	return FromSource(src)
}

func (e *Executor[M]) Pipe(operators ...sqlpipe.SourceOperator[M]) sqlpipe.Source[M] {
	return e.source().Pipe(operators...)
}

func (e *Executor[M]) IsNil() bool {
	return e.source().IsNil()
}

func (e *Executor[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return e.source().Frag(ctx)
}

func (e *Executor[M]) ApplyStmt(ctx context.Context, s *internal.Builder[M]) *internal.Builder[M] {
	return e.source().ApplyStmt(ctx, s)
}

func (e Executor[M]) PipeE(operators ...sqlpipe.SourceOperator[M]) SourceExecutor[M] {
	e.operators = append(e.operators, operators...)
	return &e
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
	s := e.src
	if s == nil {
		s = sqlpipe.From[M]()
	}

	if len(e.operators) > 0 {
		return s.Pipe(slices.SortedFunc(flatten(slices.Values(e.operators)), func(a sqlpipe.SourceOperator[M], b sqlpipe.SourceOperator[M]) int {
			return cmp.Compare(a.OperatorType(), b.OperatorType())
		})...)
	}

	return s
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

func (e *Executor[M]) Commit(ctx context.Context) error {
	_, err := e.session(ctx).Adapter().Exec(ctx, e.source())
	return err
}

func (e *Executor[M]) List(ctx context.Context) ([]*M, error) {
	list := make([]*M, 0)

	for item, err := range e.Item(ctx) {
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

func (e *Executor[M]) Range(ctx context.Context, fn func(m *M)) error {
	for item, err := range e.Item(ctx) {
		if err != nil {
			return err
		}
		fn(item)
	}
	return nil
}

func (e *Executor[M]) ListTo(ctx context.Context, listAdder Adder[M]) error {
	for item, err := range e.Item(ctx) {
		if err != nil {
			return err
		}
		listAdder.Add(item)
	}
	return nil
}

func (e *Executor[M]) FindOne(ctx context.Context) (*M, error) {
	for item, err := range e.PipeE(sqlpipe.Limit[M](1)).Item(ctx) {
		if err != nil {
			return nil, err
		}
		return item, err
	}
	return nil, nil
}

func (e *Executor[M]) Item(ctx context.Context) iter.Seq2[*M, error] {
	s := e.session(ctx)

	ex := e.source().Pipe(
		sqlpipe.DefaultProject[M](internal.ColumnsByStruct(new(M))),
	)

	x := scanner.RecvFunc[M](func(ctx context.Context, recv func(v *M) error) error {
		rows, err := s.Adapter().Query(internal.FlagContext.Inject(ctx, flags.ForReturning), ex)
		if err != nil {
			return err
		}
		return scanner.Scan(ctx, rows, scanner.Recv(recv))
	})

	return x.Item(ctx)
}

func (e *Executor[M]) CountTo(ctx context.Context, x *int64) error {
	s := e.session(ctx)

	ex := e.source().Pipe(
		sqlpipe.Project[M](sqlbuilder.Count()),
		exiternal.ForCount[M](),
	)

	rows, err := s.Adapter().Query(ctx, ex)
	if err != nil {
		return err
	}
	return scanner.Scan(ctx, rows, x)
}
