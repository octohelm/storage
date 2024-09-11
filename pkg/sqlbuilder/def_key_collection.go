package sqlbuilder

import (
	"iter"
	"strings"
)

type KeyPicker interface {
	K(name string) Key
}

type KeySeq interface {
	Keys() iter.Seq[Key]
}

type KeyCollection interface {
	KeyPicker
	KeySeq

	Of(t Table) KeyCollection
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

func (ks *keys) Keys() iter.Seq[Key] {
	return func(yield func(Key) bool) {
		for _, k := range ks.l {
			if !yield(k) {
				break
			}
		}
	}
}
