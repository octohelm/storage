package dal

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
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
	// Endpoint of database
	Endpoint string `flag:""`
	// auto migrate before run
	EnableMigrate bool `flag:",omitempty"`

	name   string
	tables *sqlbuilder.Tables
	db     adapter.Adapter
}

func (d *Database) SetDefaults() {
	if d.Endpoint == "" {
		cwd, _ := os.Getwd()
		d.Endpoint = fmt.Sprintf("sqlite://%s/%s.sqlite", cwd, d.name)
	}
}

func (d *Database) ApplyCatalog(name string, tables ...*sqlbuilder.Tables) {
	d.name = name
	d.tables = &sqlbuilder.Tables{}

	for i := range tables {
		tables[i].Range(func(tab sqlbuilder.Table, idx int) bool {
			d.tables.Add(tab)
			return true
		})
	}
}

func (d *Database) Init(ctx context.Context) error {
	if d.db != nil {
		return nil
	}

	u, err := url.Parse(d.Endpoint)
	if err != nil {
		return err
	}

	poolSize := 10
	maxConnDuration := 1 * time.Hour

	if v := u.Query().Get("pool_max_conns"); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		poolSize = int(i)
	} else {
		u.Query().Set("pool_max_conns", strconv.FormatInt(int64(poolSize), 10))
	}

	if v := u.Query().Get("pool_max_conn_lifetime"); v != "" {
		dur, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		maxConnDuration = dur
	} else {
		u.Query().Set("pool_max_conn_lifetime", maxConnDuration.String())
	}

	u.RawQuery = u.String()

	db, err := adapter.Open(ctx, u.String())
	if err != nil {
		return err
	}
	d.db = db

	ConfigureAdapter(d.db, poolSize, maxConnDuration)

	registerSessionCatalog(d.name, d.tables)

	return nil
}

func (d *Database) InjectContext(ctx context.Context) context.Context {
	return InjectContext(ctx, New(d.db, d.name))
}

func (d *Database) Run(ctx context.Context) error {
	if d.EnableMigrate == false {
		return nil
	}
	return migrator.Migrate(ctx, d.db, d.tables)
}
