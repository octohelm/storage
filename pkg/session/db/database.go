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

	_ "github.com/octohelm/storage/internal/sql/adapter/postgres"
	_ "github.com/octohelm/storage/internal/sql/adapter/sqlite"
)

type ReadonlyEndpoint struct {
	Endpoint Endpoint `flag:",omitzero"`

	EndpointOverrides
}

type Database struct {
	// Endpoint of database
	Endpoint Endpoint `flag:""`
	EndpointOverrides

	Readonly ReadonlyEndpoint

	// auto migrate before run
	EnableMigrate bool `flag:",omitzero"`

	name   string
	tables *sqlbuilder.Tables

	db   session.Adapter
	dbRo session.Adapter
}

func (d *Database) SetDefaults() {
	if d.Endpoint.IsZero() {
		cwd, _ := os.Getwd()

		d.Endpoint.Scheme = "sqlite"
		d.Endpoint.Path = path.Join(cwd, d.DBName()+".sqlite")
	}
}

func (d *Database) ApplyCatalog(name string, tables ...sqlbuilder.Catalog) {
	d.name = name
	d.tables = &sqlbuilder.Tables{}

	for _, t := range tables {
		for tt := range t.Tables() {
			d.tables.Add(tt)
		}
	}
}

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

func (d *Database) DBName() string {
	return cmp.Or(d.name, d.NameOverwrite, d.Endpoint.Base())
}

func (d *Database) Session() session.Session {
	if d.dbRo != nil {
		return session.NewWithReadOnly(d.db, d.dbRo, d.name)
	}
	return session.New(d.db, d.name)
}

func (d *Database) Catalog() sqlbuilder.Catalog {
	return d.tables
}

func (d *Database) InjectContext(ctx context.Context) context.Context {
	return session.InjectContext(ctx, d.Session())
}

func (d *Database) Run(ctx context.Context) error {
	if d.EnableMigrate == false {
		return nil
	}
	return migrator.Migrate(ctx, d.db, d.tables)
}
