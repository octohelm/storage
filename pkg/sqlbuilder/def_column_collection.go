package sqlbuilder

import (
	"context"
	"fmt"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

// ColumnCollectionManger 定义列集合的追加能力。
type ColumnCollectionManger interface {
	AddCol(cols ...Column)
}

// ColumnSeq 表示列序列。
type ColumnSeq interface {
	Cols() iter.Seq[Column]
}

// ColumnPicker 按名称或字段名挑选列。
type ColumnPicker interface {
	F(name string) Column
}

// ColumnCollect 把列序列收集为列集合。
func ColumnCollect(cols iter.Seq[Column]) ColumnCollection {
	newCols := &columns{}
	for c := range cols {
		newCols.AddCol(c)
	}
	return newCols
}

// ColumnCollection 表示可枚举、可查找的列集合。
type ColumnCollection interface {
	sqlfrag.Fragment

	ColumnSeq
	ColumnPicker

	Col(name string) Column
	AllCols() iter.Seq2[int, Column]

	Of(t Table) ColumnCollection
	Len() int
}

// Cols 按名称创建列集合。
func Cols(names ...string) ColumnCollection {
	cols := &columns{}
	for _, name := range names {
		cols.AddCol(Col(name))
	}
	return cols
}

// PickColsByFieldNames 按字段名从 picker 中挑选列集合。
func PickColsByFieldNames(picker ColumnPicker, names ...string) ColumnCollection {
	newCols := &columns{}
	for _, colName := range names {
		col := picker.F(colName)
		if col == nil {
			panic(fmt.Errorf("unknown column %s, %v", colName, names))
		}
		newCols.AddCol(col)
	}
	return newCols
}

type columns struct {
	l []Column
}

func (cols *columns) F(name string) (col Column) {
	for i := range cols.l {
		c := cols.l[i]
		if MatchColumn(c, name) {
			return c
		}
	}
	return nil
}

func (cols *columns) Col(name string) (col Column) {
	for i := range cols.l {
		c := cols.l[i]
		if MatchColumn(c, name) {
			return c
		}
	}
	return nil
}

func (cols *columns) Len() int {
	if cols == nil || cols.l == nil {
		return 0
	}
	return len(cols.l)
}

func (cols *columns) Cols() iter.Seq[Column] {
	return func(yield func(Column) bool) {
		for _, c := range cols.l {
			if !yield(c) {
				break
			}
		}
	}
}

func (cols *columns) AllCols() iter.Seq2[int, Column] {
	return func(yield func(int, Column) bool) {
		for i, c := range cols.l {
			if !yield(i, c) {
				break
			}
		}
	}
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

func (cols *columns) IsNil() bool {
	return cols == nil || cols.Len() == 0
}

func (cols *columns) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		for q, args := range sqlfrag.Join(",", cols.fragSeq()).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}

func (cols *columns) fragSeq() iter.Seq[sqlfrag.Fragment] {
	return func(yield func(sqlfrag.Fragment) bool) {
		for _, c := range cols.l {
			if !yield(c) {
				break
			}
		}
	}
}
