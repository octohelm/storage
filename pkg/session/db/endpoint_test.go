package db

import (
	"context"
	"net/url"
	"path/filepath"
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestEndpoint(t *testing.T) {
	endpoint, err := ParseEndpoint("postgres://user:pass@localhost:5432/app?sslmode=disable")

	Then(t, "ParseEndpoint 拆分连接串字段",
		Expect(err, Equal(error(nil))),
		Expect(endpoint.Scheme, Equal("postgres")),
		Expect(endpoint.Hostname, Equal("localhost")),
		Expect(endpoint.Port, Equal(uint16(5432))),
		Expect(endpoint.Path, Equal("/app")),
		Expect(endpoint.Username, Equal("user")),
		Expect(endpoint.Password, Equal("pass")),
		Expect(endpoint.Extra.Get("sslmode"), Equal("disable")),
		Expect(endpoint.Base(), Equal("app")),
		Expect(endpoint.Host(), Equal("localhost:5432")),
		Expect(endpoint.IsTLS(), Equal(true)),
	)

	Then(t, "Endpoint 可序列化为安全文本",
		Expect(endpoint.String(), Equal("postgres://user:pass@localhost:5432/app?sslmode=disable")),
		Expect(endpoint.SecurityString(), Equal("postgres://user:----@localhost:5432/app?sslmode=disable")),
	)

	var unmarshaled Endpoint
	Then(t, "Endpoint 支持 text unmarshal",
		ExpectDo(func() error {
			return unmarshaled.UnmarshalText([]byte("sqlite:///tmp/app.sqlite"))
		}),
	)
	Then(t, "Endpoint 反序列化后可读取字段",
		Expect(unmarshaled.IsZero(), Equal(false)),
		Expect(unmarshaled.Base(), Equal("app")),
		Expect(unmarshaled.IsTLS(), Equal(false)),
	)

	text, err := unmarshaled.MarshalText()
	Then(t, "MarshalText 输出连接串",
		Expect(err, Equal(error(nil))),
		Expect(string(text), Equal("sqlite:///tmp/app.sqlite")),
	)
}

func TestEndpointOverrides(t *testing.T) {
	endpoint := Endpoint{
		Scheme:   "postgres",
		Path:     "/origin",
		Username: "origin",
		Password: "secret",
	}

	overrides := EndpointOverrides{
		NameOverwrite:     "target",
		UsernameOverwrite: "user",
		PasswordOverwrite: "pass",
		ExtraOverwrite:    "sslmode=disable&connect_timeout=1",
	}

	Then(t, "EndpointOverrides 应用覆盖值",
		ExpectDo(func() error {
			return overrides.PatchEndpoint(&endpoint)
		}),
	)
	Then(t, "EndpointOverrides 覆盖用户名密码和额外参数",
		Expect(endpoint.Path, Equal("/target")),
		Expect(endpoint.Username, Equal("user")),
		Expect(endpoint.Password, Equal("pass")),
		Expect(endpoint.Extra, Equal(url.Values{
			"sslmode":         []string{"disable"},
			"connect_timeout": []string{"1"},
		})),
	)

	sqliteEndpoint := Endpoint{Scheme: "sqlite", Path: "/tmp/origin.sqlite"}
	Then(t, "sqlite 数据库名不通过 NameOverwrite 改写路径",
		ExpectDo(func() error {
			return (&EndpointOverrides{NameOverwrite: "target"}).PatchEndpoint(&sqliteEndpoint)
		}),
		Expect(sqliteEndpoint.Path, Equal("/tmp/origin.sqlite")),
	)

	Then(t, "非法 query 覆盖会返回错误",
		ExpectDo(func() error {
			return (&EndpointOverrides{ExtraOverwrite: "%zz"}).PatchEndpoint(&endpoint)
		}, ErrorMatch(mustRegexp("invalid URL escape"))),
	)
}

func TestDatabaseLifecycle(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	d := &Database{
		Endpoint: Endpoint{
			Scheme: "sqlite",
			Path:   filepath.Join(dir, "unit.sqlite"),
		},
	}

	table := sqlbuilder.T("t_user", sqlbuilder.Col("f_id"))
	d.ApplyCatalog("unit", &sqlbuilder.Tables{})
	d.tables.Add(table)

	Then(t, "Database 初始化 sqlite adapter 并注册 catalog",
		ExpectDo(func() error {
			return d.Init(ctx)
		}),
		Expect(d.DBName(), Equal("unit")),
		Expect(d.Catalog().Table("t_user").TableName(), Equal("t_user")),
		Expect(d.Run(ctx), Equal(error(nil))),
	)

	t.Cleanup(func() {
		if d.db != nil {
			_ = d.db.Close()
		}
	})

	Then(t, "Init 可重复调用且 Session 可注入上下文",
		ExpectDo(func() error {
			return d.Init(ctx)
		}),
		Expect(d.Session().Name(), Equal("unit")),
		Expect(FromContextName(d.InjectContext(ctx), "unit"), Equal("unit")),
	)

	defaulted := &Database{EndpointOverrides: EndpointOverrides{NameOverwrite: "fallback"}}
	defaulted.SetDefaults()
	Then(t, "SetDefaults 为缺省 endpoint 生成 sqlite 路径",
		Expect(defaulted.Endpoint.Scheme, Equal("sqlite")),
		Expect(defaulted.Endpoint.Base(), Equal("fallback")),
	)
}

func FromContextName(ctx context.Context, name string) string {
	return session.FromContext(ctx, name).Name()
}

func mustRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}
