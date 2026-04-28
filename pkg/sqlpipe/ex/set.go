package ex

import (
	"iter"
	"maps"
)

// AllRecords 按 key 顺序展开集合中的全部记录。
func AllRecords[ID comparable, Record any, S Set[ID, Record]](set S) iter.Seq[*Record] {
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

// FirstRecords 为每个 key 只产出第一条记录。
func FirstRecords[ID comparable, Record any, S Set[ID, Record]](set S) iter.Seq[*Record] {
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

// Set 定义按标识收集与读取记录的集合接口。
type Set[ID comparable, Record any] interface {
	Record(id ID, r *Record)

	IsZero() bool
	Keys() iter.Seq[ID]
	Records(id ID) iter.Seq[*Record]
}

// RelCache 是旧的集合别名。
// Deprecated: 请改用 Set。
type RelCache[ID comparable, Record any] = Set[ID, Record]

// OneToMulti 表示一对多记录集合。
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

// FillWith 会把指定 key 的记录逐条传给回调。
// Deprecated: 请改用 Records。
func (m OneToMulti[ID, Record]) FillWith(id ID, do func(p *Record)) {
	if list, ok := m[id]; ok {
		for _, x := range list {
			do(x)
		}
	}
}

// OneToOne 表示一对一记录集合。
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

// AllRecords 返回集合中的全部记录。
// Deprecated: 请改用包级 AllRecords。
func (m OneToOne[ID, Record]) AllRecords() iter.Seq[*Record] {
	return func(yield func(*Record) bool) {
		for _, x := range m {
			if !yield(x) {
				return
			}
		}
	}
}

// FillWith 把指定 key 的记录传给回调。
// Deprecated: 请改用 Records。
func (m OneToOne[ID, Record]) FillWith(id ID, do func(p *Record)) {
	if x, ok := m[id]; ok {
		do(x)
	}
}
