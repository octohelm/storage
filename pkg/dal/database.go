package dal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/migrator"
	"github.com/octohelm/storage/pkg/sqlbuilder"

	_ "github.com/octohelm/storage/internal/sql/adapter/postgres"
	_ "github.com/octohelm/storage/internal/sql/adapter/sqlite"
)

func ConfigureAdapter(a adapter.Adapter, poolSize int, maxConnDur time.Duration) {
	if setting, ok := a.(adapter.DBSetting); ok {
		setting.SetMaxOpenConns(poolSize)
		setting.SetMaxIdleConns(poolSize / 2)
		setting.SetConnMaxLifetime(maxConnDur)
	}
}

type Database struct {
	// Name of database key
	Name     string          `env:""`
	Endpoint string          `env:""`
	db       adapter.Adapter `env:"-"`
}

func (d *Database) SetDefaults() {
	if d.Name == "" {
		d.Name = "db"
	}

	if d.Endpoint == "" {
		cwd, _ := os.Getwd()
		d.Endpoint = fmt.Sprintf("sqlite://%s/%s.sqlite", cwd, d.Name)
	}
}

func (d *Database) Init() {
	db, err := adapter.Open(context.Background(), d.Endpoint)
	if err != nil {
		panic(err)
	}
	d.db = db
	ConfigureAdapter(d.db, 10, 1*time.Hour)
}

func (d *Database) Migrate(ctx context.Context, tables *sqlbuilder.Tables) error {
	adt := FromContext(ctx, d.Name).Adapter()
	return migrator.Migrate(ctx, adt, tables)
}

func (d *Database) InjectContext(ctx context.Context) context.Context {
	return InjectContext(ctx, New(d.db, d.Name))
}
