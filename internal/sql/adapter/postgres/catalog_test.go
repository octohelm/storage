package postgres

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestIndexSchema_ToKey(t *testing.T) {
	i := indexSchema{
		TABLE_SCHEMA: "public",
		TABLE_NAME:   "t_kubepkg_version",
		INDEX_NAME:   "t_kubepkg_version_i_version",
		INDEX_DEF:    "CREATE UNIQUE INDEX t_kubepkg_version_i_version ON public.t_kubepkg_version USING btree (f_channel, f_version)",
	}
	_ = i.ToKey(sqlbuilder.T("t_kubepkg_version"))
}
