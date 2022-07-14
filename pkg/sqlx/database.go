package sqlx

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net/url"
	"os"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"

	_ "github.com/octohelm/storage/internal/sql/adapter/postgres"
	_ "github.com/octohelm/storage/internal/sql/adapter/sqlite"
)

func init() {
	if os.Getenv("GOENV") == "DEV" {
		fmt.Println("Deprecated github.com/octohelm/storage/pkg/sqlx, to use github.com/octohelm/storage/pkg/sq instead")
	}
}

func NewFeatureDatabase(name string) *Database {
	if projectFeature, exists := os.LookupEnv("PROJECT_FEATURE"); exists && projectFeature != "" {
		name = name + "__" + projectFeature
	}
	return NewDatabase(name)
}

func NewDatabase(name string) *Database {
	return &Database{
		Name:   name,
		Tables: sqlbuilder.Tables{},
	}
}

type Database struct {
	Name   string
	Tables sqlbuilder.Tables
}

type DBNameBinder interface {
	WithDBName(dbName string) driver.Connector
}

func (database *Database) AddTable(table sqlbuilder.Table) {
	database.Tables.Add(table)
}

func (database *Database) Register(model sqlbuilder.Model) sqlbuilder.Table {
	table := sqlbuilder.TableFromModel(model)
	database.AddTable(table)
	return table
}

func (database *Database) Table(tableName string) sqlbuilder.Table {
	return database.Tables.Table(tableName)
}

func (database *Database) T(model sqlbuilder.Model) sqlbuilder.Table {
	if td, ok := model.(sqlbuilder.TableDefinition); ok {
		return td.T()
	}
	if t, ok := model.(sqlbuilder.Table); ok {
		return t
	}

	return database.Table(model.TableName())
}

func (database *Database) OpenDB(ctx context.Context, dsn string) (DBExecutor, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	if u.Path == "" || u.Path == "/" {
		u.Path = "/" + database.Name
	}

	a, err := adapter.Open(ctx, u.String())
	if err != nil {
		return nil, err
	}

	return &db{Adapter: a, db: database}, nil
}
