package db

import (
	"cmp"
	"context"
	"net/url"
	"os"
	"path"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/migrator"
	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

import (
	_ "github.com/octohelm/storage/internal/sql/adapter/postgres"
	_ "github.com/octohelm/storage/internal/sql/adapter/sqlite"
)

// ReadonlyEndpoint 描述只读数据库端点及其覆盖配置。
type ReadonlyEndpoint struct {
	Endpoint Endpoint `flag:",omitzero"`

	EndpointOverrides
}

// Database 描述一个可初始化、可注入上下文的数据库配置。
type Database struct {
	// Endpoint 是主数据库连接端点。
	Endpoint Endpoint `flag:""`
	EndpointOverrides

	Readonly ReadonlyEndpoint

	// EnableMigrate 表示启动前自动执行迁移。
	EnableMigrate bool `flag:",omitzero"`

	name   string
	tables *sqlbuilder.Tables

	db   session.Adapter
	dbRo session.Adapter
}

// SetDefaults 为缺省数据库补齐默认值。
func (d *Database) SetDefaults() {
	if d.Endpoint.IsZero() {
		cwd, _ := os.Getwd()

		d.Endpoint.Scheme = "sqlite"
		d.Endpoint.Path = path.Join(cwd, d.DBName()+".sqlite")
	}
}

// ApplyCatalog 为数据库绑定逻辑名与表目录。
func (d *Database) ApplyCatalog(name string, tables ...sqlbuilder.Catalog) {
	d.name = name
	d.tables = &sqlbuilder.Tables{}

	for _, t := range tables {
		for tt := range t.Tables() {
			d.tables.Add(tt)
		}
	}
}

// Init 初始化数据库连接、只读连接与目录注册。
func (d *Database) Init(ctx context.Context) error {
	if d.db != nil {
		return nil
	}

	endpoint := d.Endpoint

	if err := d.EndpointOverrides.PatchEndpoint(&endpoint); err != nil {
		return err
	}

	db, err := session.Open(ctx, endpoint.String())
	if err != nil {
		return err
	}

	d.db = db

	if !d.Readonly.Endpoint.IsZero() {
		readOnlyEndpoint := d.Readonly.Endpoint

		// reuse main db username & password
		if readOnlyEndpoint.Username == "" {
			if err := (&EndpointOverrides{
				UsernameOverwrite: endpoint.Username,
				PasswordOverwrite: endpoint.Password,
			}).PatchEndpoint(&readOnlyEndpoint); err != nil {
				return err
			}
		}

		if err := d.Readonly.EndpointOverrides.PatchEndpoint(&readOnlyEndpoint); err != nil {
			return err
		}

		if readOnlyEndpoint.Extra == nil {
			readOnlyEndpoint.Extra = url.Values{}
		}

		readOnlyEndpoint.Extra.Set("_ro", "true")

		dbRo, err := adapter.Open(ctx, readOnlyEndpoint.String())
		if err != nil {
			return err
		}

		d.dbRo = dbRo
	}

	session.RegisterCatalog(d.name, d.tables)

	return nil
}

// DBName 返回数据库逻辑名。
func (d *Database) DBName() string {
	return cmp.Or(d.name, d.NameOverwrite, d.Endpoint.Base())
}

// Session 返回数据库对应的会话对象。
func (d *Database) Session() session.Session {
	if d.dbRo != nil {
		return session.NewWithReadOnly(d.db, d.dbRo, d.name)
	}
	return session.New(d.db, d.name)
}

// Catalog 返回数据库绑定的表目录。
func (d *Database) Catalog() sqlbuilder.Catalog {
	return d.tables
}

// InjectContext 把数据库会话注入 context。
func (d *Database) InjectContext(ctx context.Context) context.Context {
	return session.InjectContext(ctx, d.Session())
}

// Run 按配置决定是否执行迁移。
func (d *Database) Run(ctx context.Context) error {
	if d.EnableMigrate == false {
		return nil
	}
	return migrator.Migrate(ctx, d.db, d.tables)
}
