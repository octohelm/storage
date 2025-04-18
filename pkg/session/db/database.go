package db

import (
	"cmp"
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/migrator"
	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"

	_ "github.com/octohelm/storage/internal/sql/adapter/postgres"
	_ "github.com/octohelm/storage/internal/sql/adapter/sqlite"
)

type Database struct {
	// Endpoint of database
	Endpoint Endpoint `flag:""`
	// Overwrite dbname when not empty
	NameOverwrite string `flag:",omitempty"`
	// Overwrite extra when not empty
	ExtraOverwrite string `flag:",omitempty"`
	// Overwrite username when not empty
	UsernameOverwrite string `flag:",omitempty"`
	// Overwrite password when not empty
	PasswordOverwrite string `flag:",omitempty,secret"`

	// auto migrate before run
	EnableMigrate bool `flag:",omitempty"`

	name   string
	tables *sqlbuilder.Tables
	db     adapter.Adapter
}

func (d *Database) SetDefaults() {
	if d.Endpoint.IsZero() {
		cwd, _ := os.Getwd()
		end, _ := ParseEndpoint(fmt.Sprintf("sqlite://%s/%s.sqlite", cwd, d.DBName()))
		if end != nil {
			d.Endpoint = *end
		}
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

	if name := d.NameOverwrite; name != "" {
		if endpoint.Scheme != "sqlite" {
			endpoint.Path = "/" + name
		}
	}

	if extra := d.ExtraOverwrite; extra != "" {
		q, err := url.ParseQuery(extra)
		if err != nil {
			return err
		}
		endpoint.Extra = q
	}

	if username := d.UsernameOverwrite; username != "" {
		endpoint.Username = username
	}

	if password := d.PasswordOverwrite; password != "" {
		endpoint.Password = password
	}

	db, err := adapter.Open(ctx, endpoint.String())
	if err != nil {
		return err
	}

	d.db = db

	session.RegisterCatalog(d.name, d.tables)

	return nil
}

func (d *Database) DBName() string {
	return cmp.Or(d.name, d.NameOverwrite, d.Endpoint.Base())
}

func (d *Database) InjectContext(ctx context.Context) context.Context {
	return session.InjectContext(ctx, session.New(d.db, d.name))
}

func (d *Database) Session() session.Session {
	return session.New(d.db, d.name)
}

func (d *Database) Catalog() sqlbuilder.Catalog {
	return d.tables
}

func (d *Database) Run(ctx context.Context) error {
	if d.EnableMigrate == false {
		return nil
	}
	return migrator.Migrate(ctx, d.db, d.tables)
}
