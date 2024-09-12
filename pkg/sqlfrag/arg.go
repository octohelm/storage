package sqlfrag

import (
	"context"
	"database/sql"
	"database/sql/driver"
	reflectx "github.com/octohelm/x/reflect"
	"iter"
	"reflect"
	"slices"
	"strings"
)

type NamedArg = sql.NamedArg

type NamedArgSet map[string]any

// CustomValueArg
// replace ? as some query snippet
//
// examples:
// ? => ST_GeomFromText(?)
type CustomValueArg interface {
	ValueEx() string
}

type Values[T any] iter.Seq[T]

func (seq Values[T]) IsNil() bool {
	return seq == nil
}

func (seq Values[T]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		i := 0

		for v := range seq {
			if i == 0 {
				if !yield("?", []any{v}) {
					return
				}
			} else {
				if !yield(",?", []any{v}) {
					return
				}
			}

			i++
		}
	}
}

func iterArg(ctx context.Context, v any) iter.Seq2[string, []any] {
	switch arg := v.(type) {
	case CustomValueArg:
		return func(yield func(string, []any) bool) {
			if !yield(arg.ValueEx(), []any{arg}) {
				return
			}
		}
	case Fragment:
		if !IsNil(arg) {
			return func(yield func(string, []any) bool) {
				for q, args := range arg.Frag(ctx) {
					if !yield(q, args) {
						return
					}
				}
			}
		}
	case driver.Valuer:
		return func(yield func(string, []any) bool) {
			if !yield("?", []any{arg}) {
				return
			}
		}
	case iter.Seq[any]:
		return Values[any](arg).Frag(ctx)
	case []any:
		if len(arg) > 0 {
			return Pair(strings.Repeat(",?", len(arg))[1:], arg...).Frag(ctx)
		}
	default:
		typ := reflect.TypeOf(arg)
		switch typ.Kind() {
		case reflect.Slice:
			if typ.Kind() == reflect.Slice && !reflectx.IsBytes(typ) {
				return iterArgs(ctx, arg)
			}
		case reflect.Func:
			// iter.Seq
			if typ.CanSeq() {
				rv := reflect.ValueOf(arg)

				return Values[any](func(yield func(any) bool) {
					for v := range rv.Seq() {
						if !yield(v.Interface()) {
							return
						}
					}
				}).Frag(ctx)
			}
		default:

		}

		return func(yield func(string, []any) bool) {
			if !yield("?", []any{arg}) {
				return
			}
		}
	}

	return func(yield func(string, []any) bool) {

	}
}

func iterArgs(ctx context.Context, arg any) iter.Seq2[string, []any] {
	switch x := (arg).(type) {
	case []bool:
		return Values[bool](slices.Values(x)).Frag(ctx)
	case []string:
		return Values[string](slices.Values(x)).Frag(ctx)
	case []float32:
		return Values[float32](slices.Values(x)).Frag(ctx)
	case []float64:
		return Values[float64](slices.Values(x)).Frag(ctx)
	case []int:
		return Values[int](slices.Values(x)).Frag(ctx)
	case []int8:
		return Values[int8](slices.Values(x)).Frag(ctx)
	case []int16:
		return Values[int16](slices.Values(x)).Frag(ctx)
	case []int32:
		return Values[int32](slices.Values(x)).Frag(ctx)
	case []int64:
		return Values[int64](slices.Values(x)).Frag(ctx)
	case []uint:
		return Values[uint](slices.Values(x)).Frag(ctx)
	case []uint8:
		return Values[uint8](slices.Values(x)).Frag(ctx)
	case []uint16:
		return Values[uint16](slices.Values(x)).Frag(ctx)
	case []uint32:
		return Values[uint32](slices.Values(x)).Frag(ctx)
	case []uint64:
		return Values[uint64](slices.Values(x)).Frag(ctx)
	case []any:
		return Values[any](slices.Values(x)).Frag(ctx)
	}

	sliceRv := reflect.ValueOf(arg)

	return Values[any](func(yield func(any) bool) {
		for i := 0; i < sliceRv.Len(); i++ {
			if !yield(sliceRv.Index(i).Interface()) {
				return
			}
		}
	}).Frag(ctx)
}
