package sqlbuilder

import (
	"iter"
	"strings"
)

// KeyPicker 按名称挑选索引。
type KeyPicker interface {
	K(name string) Key
}

// KeySeq 表示索引序列。
type KeySeq interface {
	Keys() iter.Seq[Key]
}

// KeyCollection 表示可复制、可查找的索引集合。
type KeyCollection interface {
	KeyPicker
	KeySeq

	Of(t Table) KeyCollection
	Len() int
}

// KeyCollectionManager 定义索引集合的追加能力。
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
