package postgres

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestIndexSchema_ToKey(t *testing.T) {
	t.Run("index from pg_indexes with USING", func(t *testing.T) {
		i := indexSchema{
			TABLE_SCHEMA: "public",
			TABLE_NAME:   "t_kubepkg_version",
			INDEX_NAME:   "t_kubepkg_version_i_version",
			INDEX_DEF:    "CREATE UNIQUE INDEX t_kubepkg_version_i_version ON public.t_kubepkg_version USING btree (f_channel, f_version)",
		}
		k := i.ToKey(sqlbuilder.T("t_kubepkg_version"))
		if k == nil {
			t.Fatal("expected a key, got nil")
		}
		if !k.IsUnique() {
			t.Fatal("expected a unique key")
		}
	})

	t.Run("constraints without USING are skipped", func(t *testing.T) {
		cases := map[string]string{
			"unique":                    "UNIQUE (f_channel, f_version)",
			"unique_nulls_not_distinct": "UNIQUE NULLS NOT DISTINCT (f_channel, f_version)",
			"foreign_key":               "FOREIGN KEY (f_id) REFERENCES t_other(f_id)",
			"check":                     "CHECK ((f_version > 0))",
		}
		for name, indexDef := range cases {
			t.Run(name, func(t *testing.T) {
				i := indexSchema{
					TABLE_SCHEMA: "public",
					TABLE_NAME:   "t_kubepkg_version",
					INDEX_NAME:   "t_kubepkg_version_i_version",
					INDEX_DEF:    indexDef,
				}
				if k := i.ToKey(sqlbuilder.T("t_kubepkg_version")); k != nil {
					t.Fatalf("expected no key, got %#v", k)
				}
			})
		}
	})

	t.Run("primary key constraint", func(t *testing.T) {
		i := indexSchema{
			TABLE_SCHEMA: "public",
			TABLE_NAME:   "t_kubepkg_version",
			INDEX_NAME:   "t_kubepkg_version_pkey",
			INDEX_DEF:    "PRIMARY KEY (f_id)",
		}
		k := i.ToKey(sqlbuilder.T("t_kubepkg_version"))
		if k == nil {
			t.Fatal("expected a key, got nil")
		}
		if !k.IsUnique() {
			t.Fatal("expected a unique key")
		}
		if k.Name() != "primary" {
			t.Fatalf("expected name primary, got %s", k.Name())
		}
	})

	t.Run("pg18 not null constraint is skipped", func(t *testing.T) {
		i := indexSchema{
			TABLE_SCHEMA: "public",
			TABLE_NAME:   "t_kubepkg_version",
			INDEX_NAME:   "t_kubepkg_version_f_name_not_null",
			INDEX_DEF:    "NOT NULL (f_name)",
		}
		if k := i.ToKey(sqlbuilder.T("t_kubepkg_version")); k != nil {
			t.Fatalf("expected no key, got %#v", k)
		}
	})
}
