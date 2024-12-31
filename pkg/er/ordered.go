package er

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"strconv"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/x/container/list"
)

type OrderedDatabase struct {
	Head

	Tables record[string, *OrderedTable] `json:"tables"`
}

func (d *OrderedDatabase) Er() *OrderedDatabase {
	return d
}

func (d *OrderedDatabase) OneOf() []any {
	return []any{
		&Database{},
	}
}

type OrderedTable struct {
	Head

	Columns     record[string, *OrderedColumn]     `json:"columns"`
	Constraints record[string, *OrderedConstraint] `json:"constraints"`
}

type OrderedColumn struct {
	Head

	Type string `json:"type"`
	Of   string `json:"of,omitzero"`

	GoType string `json:"-"`
}

type OrderedConstraint struct {
	Head

	ColumnNames []string `json:"columnNames"`
	Method      string   `json:"method,omitzero"`
	Unique      bool     `json:"unique,omitzero"`
	Primary     bool     `json:"primary,omitzero"`
}

type record[K string, V any] struct {
	props   map[K]*list.Element[*field[K, V]]
	ll      list.List[*field[K, V]]
	created bool
}

type field[K string, V any] struct {
	key   K
	value V
}

func (p *record[K, V]) initOnce() {
	if !p.created {
		p.created = true

		p.props = map[K]*list.Element[*field[K, V]]{}
		p.ll.Init()
	}
}

func (p *record[K, V]) Set(key K, value V) bool {
	p.initOnce()

	_, alreadyExist := p.props[key]
	if alreadyExist {
		p.props[key].Value.value = value
		return false
	}

	element := &field[K, V]{key: key, value: value}
	p.props[key] = p.ll.PushBack(element)
	return true
}

func (m *record[K, V]) KeyValues() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for el := m.ll.Front(); el != nil; el = el.Next() {
			if !yield(el.Value.key, el.Value.value) {
				return
			}
		}
	}
}

func (m *record[K, V]) IsZero() bool {
	return m.ll.Len() == 0
}

func (m *record[K, V]) Len() int {
	return m.ll.Len()
}

func (r *record[K, V]) Get(k K) (V, bool) {
	if r.props != nil {
		v, ok := r.props[k]
		if ok {
			return v.Value.value, true
		}
	}
	return *new(V), false
}

func (m *record[K, V]) UnmarshalJSONV2(d *jsontext.Decoder, options json.Options) error {
	t, err := d.ReadToken()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	kind := t.Kind()

	if kind != '{' {
		return &json.SemanticError{
			JSONPointer: d.StackPointer(),
			Err:         fmt.Errorf("object should starts with `{`, but got `%s`", kind),
		}
	}

	if m == nil {
		*m = record[K, V]{}
	}

	for kind := d.PeekKind(); kind != '}'; kind = d.PeekKind() {
		k, err := d.ReadValue()
		if err != nil {
			return err
		}

		key, err := strconv.Unquote(string(k))
		if err != nil {
			return &json.SemanticError{
				JSONPointer: d.StackPointer(),
				Err:         errors.New("key should be quoted string"),
			}
		}

		var v V
		if err := json.UnmarshalDecode(d, &v); err != nil {
			return err
		}
		m.Set(K(key), v)
	}

	// read the close '}'
	if _, err := d.ReadToken(); err != nil {
		if err != io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func (p record[K, V]) MarshalJSONV2(encoder *jsontext.Encoder, options json.Options) error {
	if err := encoder.WriteToken(jsontext.ObjectStart); err != nil {
		return err
	}

	for name, s := range p.KeyValues() {
		if err := json.MarshalEncode(encoder, name); err != nil {
			return err
		}

		if err := json.MarshalEncode(encoder, s); err != nil {
			return err
		}
	}

	if err := encoder.WriteToken(jsontext.ObjectEnd); err != nil {
		return err
	}

	return nil
}
