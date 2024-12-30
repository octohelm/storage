package extractor_test

import (
	"context"
	"testing"

	"github.com/octohelm/storage/pkg/er/extractor"
	"github.com/octohelm/storage/pkg/session"
	sessiondb "github.com/octohelm/storage/pkg/session/db"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/testdata/model"
	testingx "github.com/octohelm/x/testing"
)

func TestFromCatalog(t *testing.T) {
	tables := &sqlbuilder.Tables{}
	tables.Add(model.UserT)
	tables.Add(model.OrgT)
	tables.Add(model.OrgUserT)

	db := &sessiondb.Database{
		EnableMigrate: true,
	}
	db.ApplyCatalog("test", tables)

	db.SetDefaults()
	_ = db.Init(context.Background())

	ctx := db.InjectContext(context.Background())

	d := extractor.FromCatalog(ctx, session.FromContext(ctx, "test"), tables)

	s := testingx.NewSnapshot().With("er.json", testingx.MustAsJSON(d))

	testingx.Expect(t, s, testingx.MatchSnapshot("er"))
}
