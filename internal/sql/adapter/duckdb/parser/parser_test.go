package parser

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/octohelm/x/testing/bdd"
)

func TestParse(t *testing.T) {
	x := bdd.Must(Parse([]byte(`
CREATE TABLE t_user(
    f_id INTEGER DEFAULT(nextval('seq_t_user')) PRIMARY KEY, 
    f_name VARCHAR DEFAULT('') NOT NULL, 
    f_created_at TIMESTAMP DEFAULT(CURRENT_TIMESTAMP) NOT NULL, 
    f_updated_at BIGINT DEFAULT(0) NOT NULL
);
`)))

	spew.Dump(x)
}
