package sqlfrag

import (
	"context"
	"iter"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func collectSeq(seq iter.Seq2[string, []any]) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)

	for q, a := range seq {
		query.WriteString(q)
		args = append(args, a...)
	}

	return query.String(), args
}

func TestIterArgsTypedSlices(t *testing.T) {
	q, args := collectSeq(iterArgs(context.Background(), []bool{true, false}))
	qs, argss := collectSeq(iterArgs(context.Background(), []string{"a", "b"}))
	qf32, argsf32 := collectSeq(iterArgs(context.Background(), []float32{1, 2}))
	q2, args2 := collectSeq(iterArgs(context.Background(), []float64{1.5, 2.5}))
	qi, argsi := collectSeq(iterArgs(context.Background(), []int{1, 2}))
	qi8, argsi8 := collectSeq(iterArgs(context.Background(), []int8{1, 2}))
	qi16, argsi16 := collectSeq(iterArgs(context.Background(), []int16{1, 2}))
	qi32, argsi32 := collectSeq(iterArgs(context.Background(), []int32{1, 2}))
	qi64, argsi64 := collectSeq(iterArgs(context.Background(), []int64{1, 2}))
	qu, argsu := collectSeq(iterArgs(context.Background(), []uint{1, 2}))
	qu8, argsu8 := collectSeq(iterArgs(context.Background(), []uint8{1, 2}))
	q3, args3 := collectSeq(iterArgs(context.Background(), []uint16{1, 2}))
	qu32, argsu32 := collectSeq(iterArgs(context.Background(), []uint32{1, 2}))
	qu64, argsu64 := collectSeq(iterArgs(context.Background(), []uint64{1, 2}))
	qa, argsa := collectSeq(iterArgs(context.Background(), []any{"a", 1}))
	q4, args4 := collectSeq(iterArgs(context.Background(), []struct{ N int }{{1}, {2}}))

	Then(
		t, "iterArgs 支持多种切片类型",
		Expect(q, Equal("?,?")),
		Expect(args, Equal([]any{true, false})),
		Expect(qs, Equal("?,?")),
		Expect(argss, Equal([]any{"a", "b"})),
		Expect(qf32, Equal("?,?")),
		Expect(argsf32, Equal([]any{float32(1), float32(2)})),
		Expect(q2, Equal("?,?")),
		Expect(args2, Equal([]any{1.5, 2.5})),
		Expect(qi, Equal("?,?")),
		Expect(argsi, Equal([]any{1, 2})),
		Expect(qi8, Equal("?,?")),
		Expect(argsi8, Equal([]any{int8(1), int8(2)})),
		Expect(qi16, Equal("?,?")),
		Expect(argsi16, Equal([]any{int16(1), int16(2)})),
		Expect(qi32, Equal("?,?")),
		Expect(argsi32, Equal([]any{int32(1), int32(2)})),
		Expect(qi64, Equal("?,?")),
		Expect(argsi64, Equal([]any{int64(1), int64(2)})),
		Expect(qu, Equal("?,?")),
		Expect(argsu, Equal([]any{uint(1), uint(2)})),
		Expect(qu8, Equal("?,?")),
		Expect(argsu8, Equal([]any{uint8(1), uint8(2)})),
		Expect(q3, Equal("?,?")),
		Expect(args3, Equal([]any{uint16(1), uint16(2)})),
		Expect(qu32, Equal("?,?")),
		Expect(argsu32, Equal([]any{uint32(1), uint32(2)})),
		Expect(qu64, Equal("?,?")),
		Expect(argsu64, Equal([]any{uint64(1), uint64(2)})),
		Expect(qa, Equal("?,?")),
		Expect(argsa, Equal([]any{"a", 1})),
		Expect(q4, Equal("?,?")),
		Expect(len(args4), Equal(2)),
	)
}

func TestIterArgFallbacks(t *testing.T) {
	q, args := collectSeq(iterArg(context.Background(), func(yield func(int) bool) {
		yield(1)
		yield(2)
	}))
	q2, args2 := collectSeq(iterArg(context.Background(), Empty()))
	q3, args3 := collectSeq(iterArg(context.Background(), 1))
	q4, args4 := collectSeq(iterArg(context.Background(), []any{"a", 1}))
	q5, args5 := collectSeq(iterArg(context.Background(), []any{}))
	q6, args6 := collectSeq(iterArg(context.Background(), []byte("ab")))

	Then(
		t, "iterArg 支持 reflect.Seq、默认值、[]any 和空 fragment 输出",
		Expect(q, Equal("?,?")),
		Expect(args, Equal([]any{1, 2})),
		Expect(q2, Equal("")),
		Expect(args2, Equal([]any{})),
		Expect(q3, Equal("?")),
		Expect(args3, Equal([]any{1})),
		Expect(q4, Equal("?,?")),
		Expect(args4, Equal([]any{"a", 1})),
		Expect(q5, Equal("")),
		Expect(args5, Equal([]any{})),
		Expect(q6, Equal("?")),
		Expect(args6, Equal([]any{[]byte("ab")})),
	)
}
