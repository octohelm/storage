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

// SourceExecutor 表示带执行能力的数据源。
type SourceExecutor[M sqlpipe.Model] interface {
	sqlpipe.Source[M]

	// From 指定一个用于插入来源的上游数据源。
	From(source sqlpipe.Source[M]) SourceExecutor[M]
	// PipeE 以可执行数据源形式继续组合操作符。
	PipeE(operators ...sqlpipe.SourceOperator[M]) SourceExecutor[M]
	// Commit 直接执行当前数据源。
	Commit(ctx context.Context) error

	// Items 执行查询并按迭代器返回模型或错误。
	Items(ctx context.Context) iter.Seq2[*M, error]
	// FindOne 执行查询并返回一条结果；未命中时返回 nil。
	FindOne(ctx context.Context) (*M, error)
	// List 执行查询并返回结果列表。
	List(ctx context.Context) ([]*M, error)
	// ListTo 执行查询并把结果逐条写入接收器。
	ListTo(ctx context.Context, adder Adder[M]) error
	// CountTo 执行计数查询并写入目标值。
	CountTo(ctx context.Context, x *int64) error
}

// Adder 定义列表结果的接收器。
type Adder[M sqlpipe.Model] interface {
	Add(m *M)
}

// FromSource 把普通 Source 包装为可执行数据源。
func FromSource[M sqlpipe.Model](src sqlpipe.Source[M]) SourceExecutor[M] {
	return &Executor[M]{
		src: src,
	}
}

// Executor 负责组合操作符并执行数据源。
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

// IsNil 返回当前执行器对应的数据源是否可在 SQL 构建时被忽略。
func (e *Executor[M]) IsNil() bool {
	return e.source().IsNil()
}

// Frag 返回当前执行器的数据源片段。
func (e *Executor[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return e.source().Frag(ctx)
}

// ApplyStmt 把当前执行器的数据源应用到构建器。
func (e *Executor[M]) ApplyStmt(ctx context.Context, s *internal.Builder[M]) *internal.Builder[M] {
	return e.source().ApplyStmt(ctx, s)
}

// Tx 在当前会话上开启事务并执行回调。
func (e *Executor[M]) Tx(ctx context.Context, do func(ctx context.Context) error) error {
	return e.session(ctx).Tx(ctx, do)
}

// From 指定一个新的上游数据源，并返回新的执行器。
func (e *Executor[M]) From(src sqlpipe.Source[M]) SourceExecutor[M] {
	return FromSource(src)
}

// Pipe 以普通 Source 形式继续组合操作符。
func (e *Executor[M]) Pipe(operators ...sqlpipe.SourceOperator[M]) sqlpipe.Source[M] {
	return e.source().Pipe(operators...)
}

// PipeE 以执行器形式继续组合操作符。
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

// Commit 执行当前 SQL。
func (e *Executor[M]) Commit(ctx context.Context) error {
	// always for mutating
	e.forCommit = true
	_, err := e.adapterOf(ctx, e.session(ctx)).Exec(ctx, e.source())
	return err
}

// Items 执行 SQL，并以 iter.Seq2 返回模型或错误。
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

// List 执行 SQL，并返回模型切片。
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

// CountTo 执行计数 SQL，并把结果写入 int64 指针。
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

// ListTo 执行 SQL，并把模型逐条写入接收器。
func (e *Executor[M]) ListTo(ctx context.Context, listAdder Adder[M]) error {
	for item, err := range e.Items(ctx) {
		if err != nil {
			return err
		}
		listAdder.Add(item)
	}
	return nil
}

// FindOne 执行 SQL，并返回第一条结果；未命中时返回 nil。
func (e *Executor[M]) FindOne(ctx context.Context) (*M, error) {
	for item, err := range e.PipeE(sqlpipe.Limit[M](1)).Items(ctx) {
		if err != nil {
			return nil, err
		}
		return item, err
	}
	return nil, nil
}
