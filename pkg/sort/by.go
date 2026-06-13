// Package sort 提供基于枚举类型的排序选项 [By]，可生成排序方向（正序/逆序）的可枚举列表，
// 适用于 API 参数绑定、排序 UI 选项生成等场景。
//
// [By] 本身实现了 enumeration.CanEnumValues，配合 enumeration 生态可直接作为枚举值使用。
package sort

import (
	"bytes"
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-json-experiment/json"

	"github.com/octohelm/enumeration/pkg/enumeration"
)

type Order string

const (
	Asc  Order = "asc"
	Desc Order = "desc"
)

type By[E enumeration.CanEnumValues] struct {
	Field E
	Order Order

	raw []byte
}

func (a *By[E]) IsZero() bool {
	return a == nil || len(a.raw) == 0
}

func (s By[E]) EnumValues() (values []any) {
	var e E

	for _, field := range e.EnumValues() {
		for _, order := range []Order{Asc, Desc} {
			label := ""

			if x, ok := any(field).(interface{ Label() string }); ok {
				label = x.Label()
			} else {
				label = fmt.Sprintf("%v", field)
			}

			values = append(values, &enumValue{
				value: fmt.Sprintf("%v!%s", field, order),
				label: label,
			})
		}
	}

	return
}

func (v By[E]) MarshalText() ([]byte, error) {
	if len(v.raw) != 0 {
		return []byte(v.raw[:]), nil
	}

	data, err := json.Marshal(v.Field)
	if err != nil {
		return nil, err
	}

	fieldStr := string(data)

	if len(data) > 0 && data[0] == '"' {
		fieldStr, err = strconv.Unquote(fieldStr)
		if err != nil {
			return nil, err
		}
	}

	return fmt.Appendf(nil, "%s!%s", fieldStr, string(cmp.Or(v.Order, Asc))), nil
}

func (v *By[E]) UnmarshalText(data []byte) error {
	// 空字符串保持零值
	if len(data) == 0 {
		*v = By[E]{}
		return nil
	}

	parts := bytes.SplitN(data, []byte("!"), 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid sort format: %s", data)
	}
	field, order := parts[0], parts[1]

	vv := &By[E]{}

	if err := json.Unmarshal(strconv.AppendQuote(nil, string(field)), &vv.Field); err != nil {
		return fmt.Errorf("unknown sort field: %s: %w", field, err)
	}

	switch strings.ToLower(string(order)) {
	case "asc":
		vv.Order = Asc
	case "desc":
		vv.Order = Desc
	default:
		return fmt.Errorf("unknown sort order: %s", order)
	}

	vv.raw = data[:]

	*v = *vv

	return nil
}

type enumValue struct {
	value string
	label string
}

func (v enumValue) Label() string {
	label := cmp.Or(v.label, v.value)
	if strings.HasSuffix(v.value, "!desc") {
		return fmt.Sprintf("%s逆序", label)
	}
	return fmt.Sprintf("%s正序", label)
}

func (v enumValue) MarshalText() ([]byte, error) {
	return []byte(v.value), nil
}
