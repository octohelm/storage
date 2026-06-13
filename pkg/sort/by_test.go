package sort_test

import (
	"encoding"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sort "github.com/octohelm/storage/pkg/sort"
)

type UserSortBy string

const (
	UserSortByID   UserSortBy = "user~id"
	UserSortByName UserSortBy = "user~name"
)

func (UserSortBy) EnumValues() []any {
	return []any{
		UserSortByID,
		UserSortByName,
	}
}

type labeler interface {
	Label() string
}

func TestByIsZero(t *testing.T) {
	Then(
		t, "零值 By 的 IsZero 返回 true",
		Expect(new(sort.By[UserSortBy]).IsZero(), Equal(true)),
	)

	Then(
		t, "UnmarshalText 非空文本后 IsZero 返回 false",
		ExpectMustValue(func() (bool, error) {
			var b sort.By[UserSortBy]
			if err := b.UnmarshalText([]byte("user~id!asc")); err != nil {
				return false, err
			}
			return b.IsZero(), nil
		}, Equal(false)),
	)
}

func TestByMarshalTextRoundtrip(t *testing.T) {
	Then(
		t, "文本序列化往返一致",
		ExpectMustValue(func() (string, error) {
			var b sort.By[UserSortBy]
			if err := b.UnmarshalText([]byte("user~name!desc")); err != nil {
				return "", err
			}
			raw, err := b.MarshalText()
			return string(raw), err
		}, Equal("user~name!desc")),
	)

	Then(
		t, "UnmarshalText 后 Field 被正确解析",
		ExpectMustValue(func() (UserSortBy, error) {
			var b sort.By[UserSortBy]
			if err := b.UnmarshalText([]byte("user~name!desc")); err != nil {
				return "", err
			}
			return b.Field, nil
		}, Equal(UserSortByName)),
	)

	Then(
		t, "UnmarshalText 后 Order 被正确解析",
		ExpectMustValue(func() (sort.Order, error) {
			var b sort.By[UserSortBy]
			if err := b.UnmarshalText([]byte("user~name!desc")); err != nil {
				return "", err
			}
			return b.Order, nil
		}, Equal(sort.Desc)),
	)
}

func TestByEnumValues(t *testing.T) {
	var b sort.By[UserSortBy]
	values := b.EnumValues()

	Then(
		t, "EnumValues 返回 4 个排序选项（2 字段 × 2 方向）",
		Expect(len(values), Equal(4)),
	)

	Then(
		t, "正序选项的 Label 正确",
		Expect(values[0].(labeler).Label(), Equal("user~id正序")),
		Expect(values[2].(labeler).Label(), Equal("user~name正序")),
	)

	Then(
		t, "逆序选项的 Label 正确",
		Expect(values[1].(labeler).Label(), Equal("user~id逆序")),
		Expect(values[3].(labeler).Label(), Equal("user~name逆序")),
	)

	Then(
		t, "选项序列化为 field!order 格式",
		ExpectMustValue(func() (string, error) {
			raw, err := values[0].(encoding.TextMarshaler).MarshalText()
			return string(raw), err
		}, Equal("user~id!asc")),
		ExpectMustValue(func() (string, error) {
			raw, err := values[1].(encoding.TextMarshaler).MarshalText()
			return string(raw), err
		}, Equal("user~id!desc")),
		ExpectMustValue(func() (string, error) {
			raw, err := values[2].(encoding.TextMarshaler).MarshalText()
			return string(raw), err
		}, Equal("user~name!asc")),
		ExpectMustValue(func() (string, error) {
			raw, err := values[3].(encoding.TextMarshaler).MarshalText()
			return string(raw), err
		}, Equal("user~name!desc")),
	)
}
