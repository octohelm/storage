package session

import (
	"context"
	"fmt"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

// OptionFunc 表示会话适配器选项函数。
type OptionFunc func(*option)

// ReadOnly 要求 Session.Adapter 在可用时返回只读适配器。
func ReadOnly() OptionFunc {
	return func(o *option) {
		o.ReadyOnly = true
	}
}

type option struct {
	ReadyOnly bool
}

// Session 表示数据库会话接口。
type Session interface {
	// Name 返回逻辑会话名。
	Name() string
	// T 把模型或表包装值解析为当前会话使用的表。
	T(m any) sqlbuilder.Table

	// Tx 使用可写适配器开启事务并执行 fn。
	Tx(ctx context.Context, fn func(ctx context.Context) error) error

	// Adapter 根据选项返回可写或只读适配器。
	Adapter(options ...OptionFunc) Adapter
}

// New 创建一个由单个适配器支撑的会话。
func New(a Adapter, name string) Session {
	return &session{
		name:    name,
		adapter: a,
	}
}

// NewWithReadOnly 创建一个读写适配器分离的会话。
func NewWithReadOnly(a adapter.Adapter, ro adapter.Adapter, name string) Session {
	return &session{
		name:      name,
		adapter:   a,
		adapterRo: ro,
	}
}

type session struct {
	name      string
	adapter   adapter.Adapter
	adapterRo adapter.Adapter
}

func (s *session) Adapter(optFns ...OptionFunc) adapter.Adapter {
	if s.adapterRo != nil {
		opt := &option{}
		for _, optFn := range optFns {
			optFn(opt)
		}

		if opt.ReadyOnly {
			return s.adapterRo
		}
	}

	return s.adapter
}

func (s *session) Name() string {
	return s.name
}

func (s *session) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.adapter.Transaction(ctx, fn)
}

func (s *session) T(m any) sqlbuilder.Table {
	if td, ok := m.(sqlbuilder.WithTable); ok {
		return td.T()
	}
	if td, ok := m.(sqlbuilder.Table); ok {
		return td
	}
	return sqlbuilder.TableFromModel(m)
}

// TableWrapper 表示可解包为模型的表包装。
type TableWrapper interface {
	Unwrap() sqlbuilder.Model
}

// MustFor 返回按名称或表匹配到的会话，缺失时直接 panic。
func MustFor(ctx context.Context, nameOrTable any) Session {
	s := For(ctx, nameOrTable)
	if s == nil {
		panic(fmt.Errorf("invalid section target %#v", nameOrTable))
	}
	return s
}

// For 根据名称、模型或 TableWrapper 从 context 中解析会话。
func For(ctx context.Context, nameOrTable any) Session {
	if u, ok := nameOrTable.(TableWrapper); ok {
		return For(ctx, u.Unwrap())
	}

	switch x := nameOrTable.(type) {
	case string:
		return FromContext(ctx, x)
	case sqlbuilder.Model:
		if t, ok := catalogs.Load(x.TableName()); ok {
			return FromContext(ctx, t.(string))
		}
	}

	return nil
}

type contextSession struct {
	name string
}

// InjectContext 按逻辑名称把会话注入 context。
func InjectContext(ctx context.Context, repo Session) context.Context {
	return contextx.WithValue(ctx, contextSession{name: repo.Name()}, repo)
}

// FromContext 返回 context 中指定名称的会话，缺失时直接 panic。
func FromContext(ctx context.Context, name string) Session {
	r, ok := ctx.Value(contextSession{name: name}).(Session)
	if ok {
		return r
	}
	panic(fmt.Sprintf("missing session of %s", name))
}
