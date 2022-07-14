package sqlbuilder

import (
	"fmt"
	"strings"
)

type KeyCollection interface {
	Of(t Table) KeyCollection

	K(name string) Key
	RangeKey(func(k Key, i int) bool)
	Keys(names ...string) KeyCollection
	Len() int
}

type KeyCollectionManager interface {
	AddKey(keys ...Key)
}

type keys struct {
	l []Key
}

func (ks *keys) Len() int {
	if ks == nil {
		return 0
	}
	return len(ks.l)
}

func (ks *keys) K(keyName string) Key {
	keyName = strings.ToLower(keyName)
	for i := range ks.l {
		k := ks.l[i]
		if keyName == k.Name() {
			return k
		}
	}
	return nil
}

func (ks *keys) AddKey(nextKeys ...Key) {
	for i := range nextKeys {
		k := nextKeys[i]
		if k == nil {
			continue
		}
		ks.l = append(ks.l, k)
	}
}

func (ks *keys) Of(newTable Table) KeyCollection {
	newKeys := &keys{}
	for i := range ks.l {
		newKeys.AddKey(ks.l[i].Of(newTable))
	}
	return newKeys
}

func (ks *keys) Keys(names ...string) KeyCollection {
	if len(names) == 0 {
		return &keys{
			l: ks.l,
		}
	}

	newCols := &keys{}
	for _, indexName := range names {
		col := ks.K(indexName)
		if col == nil {
			panic(fmt.Errorf("unknown index %s", indexName))
		}
		newCols.AddKey(col)
	}
	return newCols
}

func (ks *keys) RangeKey(cb func(col Key, idx int) bool) {
	for i := range ks.l {
		if !cb(ks.l[i], i) {
			break
		}
	}
}
