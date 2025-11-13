package duckdb

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/octohelm/storage/internal/testutil"
)

func Test_parseTableCreate(t *testing.T) {
	decls := parseTableDecls(bytes.NewBuffer([]byte(`
CREATE TABLE t_user(
    f_id INTEGER DEFAULT(nextval('seq_t_user')) PRIMARY KEY, 
    f_name VARCHAR DEFAULT('') NOT NULL, 
    f_created_at TIMESTAMP DEFAULT(CURRENT_TIMESTAMP) NOT NULL, 
    f_updated_at BIGINT DEFAULT(0) NOT NULL, 
);
`)))

	spew.Dump(decls)

	testutil.Expect(t, decls, testutil.Equal(map[string][]string{
		"f_id":         {"INTEGER", "DEFAULT(nextval('seq_t_user'))", "PRIMARY KEY"},
		"f_name":       {"TEXT", "DEFAULT('')", "NOT NULL"},
		"f_created_at": {"TIMESTAMP", "DEFAULT(CURRENT_TIMESTAMP)", "NOT NULL"},
		"f_updated_at": {"BIGINT", "DEFAULT(0)", "NOT NULL"},
	}))
}
