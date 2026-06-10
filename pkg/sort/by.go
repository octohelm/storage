// Package sort 提供基于枚举类型的排序选项 [By]，可生成排序方向（正序/逆序）的可枚举列表，
// 适用于 API 参数绑定、排序 UI 选项生成等场景。
//
// [By] 本身实现了 enumeration.CanEnumValues，配合 enumeration 生态可直接作为枚举值使用。
package sort

import (
	"cmp"
	"fmt"
	"strings"

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

	value string
}

func (a *By[E]) IsZero() bool {
	return a.value == ""
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
	return []byte(v.value), nil
}

func (v *By[E]) UnmarshalText(data []byte) error {
	by := By[E]{
		value: string(data),
	}

	*v = by
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
