package ex

import (
	"golang.org/x/exp/maps"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type OneToMulti[ID comparable, Record any] map[ID][]*Record

func (m OneToMulti[ID, Record]) Record(id ID, r *Record) {
	m[id] = append(m[id], r)
}

func (m OneToMulti[ID, Record]) IsZero() bool {
	return len(m) == 0
}

func (m OneToMulti[ID, Record]) Keys() []ID {
	return maps.Keys(m)
}

func (m OneToMulti[ID, Record]) FillWith(id ID, do func(p *Record)) {
	if list, ok := m[id]; ok {
		for _, x := range list {
			do(x)
		}
	}
}

func (m OneToMulti[ID, Record]) Records(id ID) iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		if list, ok := m[id]; ok {
			for _, x := range list {
				if !yield(x) {
					return
				}
			}

		}
	}
}

type OneToOne[ID comparable, Record any] map[ID]*Record

func (m OneToOne[ID, Record]) Record(id ID, r *Record) {
	m[id] = r
}

func (m OneToOne[ID, Record]) IsZero() bool {
	return len(m) == 0
}

func (m OneToOne[ID, Record]) Keys() []ID {
	return maps.Keys(m)
}

func (m OneToOne[ID, Record]) AsInKeys() sqlbuilder.ColumnValuer[ID] {
	return sqlbuilder.In(maps.Keys(m)...)
}

func (m OneToOne[ID, Record]) FillWith(id ID, do func(p *Record)) {
	if x, ok := m[id]; ok {
		do(x)
	}
}

func (m OneToOne[ID, Record]) Records(id ID) iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		if x, ok := m[id]; ok {
			if !yield(x) {
				return
			}
		}
	}
}
