package builder_test

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"
	. "github.com/octohelm/sqlx/pkg/builder"
)

func BenchmarkCols(b *testing.B) {
	columns := Columns{}

	columns.Add(
		Col("f_id").Field("ID").Type(1, `,autoincrement`),
		Col("f_name").Field("Name").Type(1, ``),
		Col("f_f1").Field("F1").Type(1, ``),
		Col("f_f2").Field("F2").Type(1, ``),
		Col("f_f3").Field("F3").Type(1, ``),
		Col("f_f4").Field("F4").Type(1, ``),
		Col("f_f5").Field("F5").Type(1, ``),
		Col("f_f6").Field("F6").Type(1, ``),
		Col("f_f7").Field("F7").Type(1, ``),
		Col("f_f8").Field("F8").Type(1, ``),
		Col("f_f9").Field("F9").Type(1, ``),
	)

	b.Run("pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = columns.F("F3")
		}
	})

	b.Run("multi pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = columns.Fields("ID", "Name")
		}
	})

	b.Run("multi pick all", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = columns.Fields()
		}
	})

}

func TestColumns(t *testing.T) {
	columns := Columns{}

	t.Run("empty columns", func(t *testing.T) {
		testutil.Expect(t, columns.Len(), testutil.Equal(0))
		testutil.Expect(t, columns.AutoIncrement(), testutil.Be[*Column](nil))
	})

	t.Run("added cols", func(t *testing.T) {
		columns.Add(
			Col("F_id").Field("ID").Type(1, `,autoincrement`),
		)

		autoIncrementCol := columns.AutoIncrement()

		testutil.Expect(t, autoIncrementCol, testutil.Not(testutil.Be[*Column](nil)))
		testutil.Expect(t, autoIncrementCol.Name, testutil.Equal("f_id"))

		t.Run("get col by FieldName", func(t *testing.T) {

			testutil.Expect(t, columns.F("ID2"), testutil.Be[*Column](nil))

			testutil.Expect(t, MustCols(columns.Fields("ID2")).Len(), testutil.Equal(0))
			testutil.Expect(t, MustCols(columns.Fields()).Len(), testutil.Equal(1))

			testutil.Expect(t, MustCols(columns.Fields("ID2")).List(), testutil.HaveLen[[]*Column](0))
			testutil.Expect(t, MustCols(columns.Fields()).Len(), testutil.Equal(1))
		})
		t.Run("get col by ColName", func(t *testing.T) {
			testutil.Expect(t, MustCols(columns.Cols("F_id")).Len(), testutil.Equal(1))
			testutil.Expect(t, MustCols(columns.Cols()).Len(), testutil.Be(1))
			testutil.Expect(t, MustCols(columns.Cols()).List(), testutil.HaveLen[[]*Column](1))

			testutil.Expect(t, MustCols(columns.Cols()).FieldNames(), testutil.Equal([]string{"ID"}))
		})
	})
}

func MustCols(cols *Columns, err error) *Columns {
	return cols
}
