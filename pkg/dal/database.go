package dal

import (
	"cmp"
	"context"
	"fmt"
	"github.com/octohelm/storage/pkg/datatypes"
	"os"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/migrator"
	"github.com/octohelm/storage/pkg/sqlbuilder"

	_ "github.com/octohelm/storage/internal/sql/adapter/postgres"
	_ "github.com/octohelm/storage/internal/sql/adapter/sqlite"
)

type Database struct {
	// Endpoint of database
	Endpoint datatypes.Endpoint `flag:""`
	// Overwrite dbname when not empty
	NameOverwrites string `flag:",omitempty"`
	// Overwrite username when not empty
	UsernameOverwrites string `flag:",omitempty"`
	// Overwrite password when not empty
	PasswordOverwrites string `flag:",omitempty,secret"`

	// auto migrate before run
	EnableMigrate bool `flag:",omitempty"`

	name   string
	tables *sqlbuilder.Tables
	db     adapter.Adapter
}

func (d *Database) SetDefaults() {
	if d.Endpoint.IsZero() {
		cwd, _ := os.Getwd()
		end, _ := datatypes.ParseEndpoint(fmt.Sprintf("sqlite://%s/%s.sqlite", cwd, d.DBName()))
		if end != nil {
			d.Endpoint = *end
		}
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

	endpoint := d.Endpoint

	if name := d.NameOverwrites; name != "" {
		endpoint.Path = "/" + name
	}

	if username := d.UsernameOverwrites; username != "" {
		endpoint.Username = username
	}

	if password := d.PasswordOverwrites; password != "" {
		endpoint.Password = password
	}

	db, err := adapter.Open(ctx, endpoint.String())
	if err != nil {
		return err
	}

	d.db = db
	registerSessionCatalog(d.name, d.tables)
	return nil
}

func (d *Database) DBName() string {
	return cmp.Or(d.name, d.Endpoint.Base())
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
