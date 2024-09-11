package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func BenchmarkCols(b *testing.B) {
	cols := Cols()

	(cols).(ColumnCollectionManger).AddCol(
		Col("f_id", ColField("ID"), ColTypeOf(1, `,autoincrement`)),
		Col("f_name", ColField("Name"), ColTypeOf(1, ``)),
		Col("f_f1", ColField("F1"), ColTypeOf(1, ``)),
		Col("f_f2", ColField("F2"), ColTypeOf(1, ``)),
		Col("f_f3", ColField("F3"), ColTypeOf(1, ``)),
		Col("f_f4", ColField("F4"), ColTypeOf(1, ``)),
		Col("f_f5", ColField("F5"), ColTypeOf(1, ``)),
		Col("f_f6", ColField("F6"), ColTypeOf(1, ``)),
		Col("f_f7", ColField("F7"), ColTypeOf(1, ``)),
		Col("f_f8", ColField("F8"), ColTypeOf(1, ``)),
		Col("f_f9", ColField("F9"), ColTypeOf(1, ``)),
	)

	b.Run("pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cols.F("F3")
		}
	})

	b.Run("multi pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = PickColsByFieldNames(cols, "ID", "Name")
		}
	})
}

func TestColumns(t *testing.T) {
	cols := Cols()

	t.Run("empty columns", func(t *testing.T) {
		testutil.Expect(t, cols.Len(), testutil.Equal(0))
	})

	t.Run("added cols", func(t *testing.T) {
		cols.(ColumnCollectionManger).AddCol(
			Col("F_id", ColField("ID"), ColTypeOf(1, `,autoincrement`)),
		)

		t.Run("get col by FieldName", func(t *testing.T) {
			testutil.Expect(t, cols.F("ID2"), testutil.Be[Column](nil))

			testutil.Expect(t, PickColsByFieldNames(cols, "ID").Len(), testutil.Equal(1))
			testutil.Expect(t, cols.Len(), testutil.Equal(1))
		})

		t.Run("get col by ColName", func(t *testing.T) {
			testutil.Expect(t, PickColsByFieldNames(cols, "f_id").Len(), testutil.Equal(1))
			testutil.Expect(t, cols.Len(), testutil.Be(1))
		})
	})
}
