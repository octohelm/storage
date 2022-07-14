package sqlbuilder

import (
	"context"
	"fmt"
)

type ColumnCollectionManger interface {
	AddCol(cols ...Column)
}

func ColumnCollectionFromList(list []Column) ColumnCollection {
	return &columns{l: list}
}

type ColumnCollection interface {
	SqlExpr

	Of(t Table) ColumnCollection

	F(name string) Column
	Col(name string) Column
	Cols(names ...string) ColumnCollection
	RangeCol(cb func(col Column, idx int) bool)
	Len() int
}

func Cols(names ...string) ColumnCollection {
	cols := &columns{}
	for _, name := range names {
		cols.AddCol(Col(name))
	}
	return cols
}

type columns struct {
	l []Column
}

func (cols *columns) IsNil() bool {
	return cols == nil || cols.Len() == 0
}

func (cols *columns) F(name string) (col Column) {
	for i := range cols.l {
		c := cols.l[i]
		if c.MatchName(name) {
			return c
		}
	}
	return nil
}

func (cols *columns) Col(name string) (col Column) {
	for i := range cols.l {
		c := cols.l[i]
		if c.MatchName(name) {
			return c
		}
	}
	return nil
}

func (cols *columns) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.Grow(cols.Len())

	cols.RangeCol(func(col Column, idx int) bool {
		if idx > 0 {
			e.WriteQueryByte(',')
		}
		e.WriteExpr(col)
		return true
	})

	return e.Ex(ctx)
}

func (cols *columns) Len() int {
	if cols == nil || cols.l == nil {
		return 0
	}
	return len(cols.l)
}

func (cols *columns) RangeCol(cb func(col Column, idx int) bool) {
	for i := range cols.l {
		if !cb(cols.l[i], i) {
			break
		}
	}
}

func (cols *columns) Cols(names ...string) ColumnCollection {
	if len(names) == 0 {
		return &columns{
			l: cols.l,
		}
	}

	newCols := &columns{}
	for _, colName := range names {
		col := cols.F(colName)
		if col == nil {
			panic(fmt.Errorf("unknown column %s", colName))
		}
		newCols.AddCol(col)
	}
	return newCols
}

func (cols *columns) Of(newTable Table) ColumnCollection {
	newCols := &columns{}
	for i := range cols.l {
		newCols.AddCol(cols.l[i].Of(newTable))
	}
	return newCols
}

func (cols *columns) AddCol(columns ...Column) {
	for i := range columns {
		col := columns[i]
		if col == nil {
			continue
		}
		cols.l = append(cols.l, col)
	}
}
