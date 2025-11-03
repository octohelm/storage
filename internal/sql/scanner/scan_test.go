package scanner

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/dberr"
)

type T struct {
	I int    `db:"f_i"`
	S string `db:"f_s"`
}

type Any string

type T2 T

func (t *T2) ColumnReceivers() map[string]interface{} {
	return map[string]interface{}{
		"f_i": &t.I,
		"f_s": &t.S,
	}
}

type TDataList struct {
	Data []T
}

func (*TDataList) New() interface{} {
	return &T{}
}

func (l *TDataList) Next(v interface{}) error {
	t := v.(*T)
	l.Data = append(l.Data, *t)
	return nil
}

func BenchmarkScan(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	b.Run("Scan to struct", func(b *testing.B) {
		sql := "SELECT f_i,f_s from t"

		mockRows := mock.NewRows([]string{"f_i", "f_s"})
		mockRows.AddRow(2, "4")

		_ = mock.ExpectQuery(sql).WillReturnRows(mockRows)

		target := &T{}

		for i := 0; i < b.N; i++ {
			rows, _ := db.Query(sql)
			_ = Scan(context.Background(), rows, target)
		}

		b.Log(target)
	})

	b.Run("Scan to struct with column receivers", func(b *testing.B) {
		sql := "SELECT f_i,f_s from t"

		mockRows := mock.NewRows([]string{"f_i", "f_s"})
		mockRows.AddRow(2, "4")

		_ = mock.ExpectQuery(sql).WillReturnRows(mockRows)

		target := &T2{}

		for i := 0; i < b.N; i++ {
			rows, _ := db.Query(sql)
			_ = Scan(context.Background(), rows, target)
		}

		b.Log(target)
	})
}

func TestScan(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	t.Run("Scan to struct", func(t *testing.T) {
		sql := "SELECT f_i,f_s from t"

		mockRows := mock.NewRows([]string{"f_i", "f_s"})
		mockRows.AddRow(2, "4")

		_ = mock.ExpectQuery(sql).WillReturnRows(mockRows)

		target := &T{}
		rows, _ := db.Query(sql)
		err := Scan(context.Background(), rows, target)
		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, target, testutil.Equal(&T{
			I: 2,
			S: "4",
		}))
	})

	t.Run("Scan to struct with column receivers", func(t *testing.T) {
		sql := "SELECT f_i,f_s from t"

		mockRows := mock.NewRows([]string{"f_i", "f_s"})
		mockRows.AddRow(2, "4")

		_ = mock.ExpectQuery(sql).WillReturnRows(mockRows)

		target := &T2{}
		rows, _ := db.Query(sql)
		err := Scan(context.Background(), rows, target)
		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, target, testutil.Equal(&T2{
			I: 2,
			S: "4",
		}))
	})

	t.Run("Scan to struct without no record", func(t *testing.T) {
		sql := "SELECT f_i,f_s from t"

		mockRows := mock.NewRows([]string{"f_i", "f_s"})

		_ = mock.ExpectQuery(sql).WillReturnRows(mockRows)

		target := &T{}
		rows, err := db.Query(sql)
		testutil.Expect(t, err, testutil.Be[error](nil))

		err = Scan(context.Background(), rows, target)
		testutil.Expect(t, dberr.IsErrNotFound(err), testutil.Be(true))
	})

	t.Run("Scan to count", func(t *testing.T) {
		mockRows := mock.NewRows([]string{"count(1)"})
		mockRows.AddRow(10)

		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		count := 0
		rows, err := db.Query("SELECT count(1) from t")
		testutil.Expect(t, err, testutil.Be[error](nil))

		err = Scan(context.Background(), rows, &count)
		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, count, testutil.Equal(10))
	})

	t.Run("Scan to count failed when bad receiver", func(t *testing.T) {
		mockRows := mock.NewRows([]string{"count(1)"})
		mockRows.AddRow(10)

		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		v := Any("")
		rows, err := db.Query("SELECT count(1) from t")
		testutil.Expect(t, err, testutil.Be[error](nil))

		err = Scan(context.Background(), rows, &v)
		testutil.Expect(t, err, testutil.Not(testutil.Be[error](nil)))
	})

	t.Run("Scan to slice", func(t *testing.T) {
		mockRows := mock.NewRows([]string{"f_i", "f_s"})
		mockRows.AddRow(2, "2")
		mockRows.AddRow(3, "3")

		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		list := make([]T, 0)
		rows, err := db.Query("SELECT f_i,f_b from t")
		testutil.Expect(t, err, testutil.Be[error](nil))

		err = Scan(context.Background(), rows, &list)

		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, list, testutil.Equal([]T{
			{
				I: 2,
				S: "2",
			},
			{
				I: 3,
				S: "3",
			},
		}))
	})

	t.Run("Scan to iterator", func(t *testing.T) {
		mockRows := mock.NewRows([]string{"f_i", "f_s"})
		mockRows.AddRow(2, "2")
		mockRows.AddRow(3, "3")

		_ = mock.ExpectQuery("SELECT .+ from t").WillReturnRows(mockRows)

		rows, err := db.Query("SELECT f_i,f_b from t")
		testutil.Expect(t, err, testutil.Be[error](nil))

		list := TDataList{}

		err = Scan(context.Background(), rows, &list)

		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, list.Data, testutil.Equal([]T{
			{
				I: 2,
				S: "2",
			},
			{
				I: 3,
				S: "3",
			},
		}))
	})
}
