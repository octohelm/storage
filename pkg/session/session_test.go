package session

import (
	"context"
	"database/sql"
	"net/url"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	internaladapter "github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

type stubAdapter struct {
	name string
	tx   int
}

func (a *stubAdapter) Exec(ctx context.Context, expr sqlfrag.Fragment) (sql.Result, error) {
	return nil, nil
}

func (a *stubAdapter) Query(ctx context.Context, expr sqlfrag.Fragment) (*sql.Rows, error) {
	return nil, nil
}
func (a *stubAdapter) Close() error                     { return nil }
func (a *stubAdapter) DriverName() string               { return a.name }
func (a *stubAdapter) Dialect() internaladapter.Dialect { return nil }
func (a *stubAdapter) Catalog(ctx context.Context) (*sqlbuilder.Tables, error) {
	return &sqlbuilder.Tables{}, nil
}

func (a *stubAdapter) Transaction(ctx context.Context, action func(ctx context.Context) error) error {
	a.tx++
	return action(ctx)
}

func (a *stubAdapter) Open(ctx context.Context, dsn *url.URL) (internaladapter.Adapter, error) {
	return a, nil
}

type tableWrapper struct{ model.User }

func (tableWrapper) Unwrap() sqlbuilder.Model { return &model.User{} }

func TestSessionBasics(t *testing.T) {
	mainAdapter := &stubAdapter{name: "main"}
	roAdapter := &stubAdapter{name: "ro"}
	s := NewWithReadOnly(mainAdapter, roAdapter, "primary")

	Then(t, "Session 返回名称和只读 adapter",
		Expect(s.Name(), Equal("primary")),
		Expect(s.Adapter(), Equal(Adapter(mainAdapter))),
		Expect(s.Adapter(ReadOnly()), Equal(Adapter(roAdapter))),
	)

	Then(t, "Session.Tx 透传到底层 adapter",
		ExpectDo(func() error {
			return s.Tx(context.Background(), func(ctx context.Context) error { return nil })
		}),
	)
	Then(t, "Session.Tx 调用底层事务",
		Expect(mainAdapter.tx, Equal(1)),
	)

	Then(t, "Session.T 支持模型和表",
		Expect(s.T(&model.User{}).TableName(), Equal("t_user")),
		Expect(s.T(model.UserT).TableName(), Equal("t_user")),
	)
}

func TestSessionContextAndRegistry(t *testing.T) {
	s := New(&stubAdapter{name: "main"}, "primary")
	ctx := InjectContext(context.Background(), s)

	catalog := &sqlbuilder.Tables{}
	catalog.Add(sqlbuilder.TableFromModel(&model.User{}))
	RegisterCatalog("primary", catalog)

	Then(t, "FromContext 和 For 根据名称与模型查找 session",
		Expect(FromContext(ctx, "primary").Name(), Equal("primary")),
		Expect(For(ctx, "primary").Name(), Equal("primary")),
		Expect(For(ctx, &model.User{}).Name(), Equal("primary")),
		Expect(For(ctx, tableWrapper{}).Name(), Equal("primary")),
	)

	Then(t, "MustFor 在目标不存在时 panic",
		ExpectMustValue(func() (panicked bool, err error) {
			defer func() { panicked = recover() != nil }()
			_ = MustFor(context.Background(), "missing")
			return panicked, nil
		}, Equal(true)),
	)
}

func TestHelpers(t *testing.T) {
	Then(t, "Recv 透传 scanner.Recv",
		Expect(Recv(func(v *int) error { return nil }) != nil, Equal(true)),
		Expect(Scan != nil, Equal(true)),
	)

	Then(t, "InTx 识别非事务上下文",
		Expect(InTx(context.Background()), Equal(false)),
	)
}
