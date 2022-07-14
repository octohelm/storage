package sqlite

import (
	"bytes"
	"testing"

	"github.com/octohelm/storage/internal/testutil"
)

func Test_parseSQL(t *testing.T) {
	cols := extractCols(bytes.NewBuffer([]byte(`
CREATE TABLE t_user (
	f_id UNSIGNED BIG INT NOT NULL,
	f_name TEXT NOT NULL DEFAULT '',
	f_nickname TEXT NOT NULL DEFAULT '',
	f_username TEXT NOT NULL DEFAULT '',
	f_gender INTEGER NOT NULL DEFAULT '0',
	f_created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	f_updated_at BIGINT NOT NULL DEFAULT '0',
	PRIMARY KEY (f_id)
)
`)))
	testutil.Expect(t, cols, testutil.Equal(map[string]string{
		"f_gender":     "INTEGER NOT NULL DEFAULT '0'",
		"f_created_at": "timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"f_updated_at": "BIGINT NOT NULL DEFAULT '0'",
		"PRIMARY":      "KEY ( f_id )",
		"f_id":         "UNSIGNED BIG INT NOT NULL",
		"f_name":       "TEXT NOT NULL DEFAULT ''",
		"f_nickname":   "TEXT NOT NULL DEFAULT ''",
		"f_username":   "TEXT NOT NULL DEFAULT ''",
	}))
}
