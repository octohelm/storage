package ex

import (
	"iter"
	"maps"
)

func AllRecords[ID comparable, Record any](set Set[ID, Record]) iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		for key := range set.Keys() {
			for v := range set.Records(key) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func FirstRecords[ID comparable, Record any](set Set[ID, Record]) iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		for key := range set.Keys() {
			for v := range set.Records(key) {
				if !yield(v) {
					return
				}
				// pick first only for each key
				break
			}
		}
	}
}

type Set[ID comparable, Record any] interface {
	Record(id ID, r *Record)

	IsZero() bool
	Keys() iter.Seq[ID]
	Records(id ID) iter.Seq[*Record]
}

// RelCache
// Deprecated use Set instead
type RelCache[ID comparable, Record any] = Set[ID, Record]

type OneToMulti[ID comparable, Record any] map[ID][]*Record

var _ Set[int, int] = OneToMulti[int, int]{}

func (m OneToMulti[ID, Record]) IsZero() bool {
	return len(m) == 0
}

func (m OneToMulti[ID, Record]) Record(id ID, r *Record) {
	m[id] = append(m[id], r)
}

func (m OneToMulti[ID, Record]) Keys() iter.Seq[ID] {
	return maps.Keys(m)
}

func (m OneToMulti[ID, Record]) AllRecords() iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		for _, list := range m {
			for _, x := range list {
				if !yield(x) {
					return
				}
			}
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

// FillWith Deprecated use Records instead
func (m OneToMulti[ID, Record]) FillWith(id ID, do func(p *Record)) {
	if list, ok := m[id]; ok {
		for _, x := range list {
			do(x)
		}
	}
}

type OneToOne[ID comparable, Record any] map[ID]*Record

var _ Set[int, int] = OneToOne[int, int]{}

func (m OneToOne[ID, Record]) Record(id ID, r *Record) {
	m[id] = r
}

func (m OneToOne[ID, Record]) IsZero() bool {
	return len(m) == 0
}

func (m OneToOne[ID, Record]) Keys() iter.Seq[ID] {
	return maps.Keys(m)
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

// AllRecords
// Deprecated use AllRecords instead
func (m OneToOne[ID, Record]) AllRecords() iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		for _, x := range m {
			if !yield(x) {
				return
			}
		}
	}
}

// FillWith  Deprecated use Records instead
func (m OneToOne[ID, Record]) FillWith(id ID, do func(p *Record)) {
	if x, ok := m[id]; ok {
		do(x)
	}
}
