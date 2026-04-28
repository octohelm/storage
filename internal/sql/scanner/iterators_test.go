package scanner

import (
	"context"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestIterators(t *testing.T) {
	values := make([]int, 0)
	iter := Recv(func(v *int) error {
		values = append(values, *v)
		return nil
	})

	item := iter.New().(*int)
	*item = 1
	Then(t, "typedScanner 创建并接收具体类型值",
		ExpectDo(func() error { return iter.Next(item) }),
	)
	Then(t, "typedScanner 回调收到扫描值",
		Expect(values, Equal([]int{1})),
	)

	var items []int
	sliceIter, err := ScanIteratorFor(&items)
	Then(t, "slice 目标创建 SliceScanIterator",
		Expect(err, Equal(error(nil))),
		ExpectDo(func() error {
			v := sliceIter.New().(*int)
			*v = 2
			return sliceIter.Next(v)
		}),
	)
	Then(t, "SliceScanIterator 追加结果到切片",
		Expect(items, Equal([]int{2})),
	)

	target := 0
	singleIter, err := ScanIteratorFor(&target)
	Then(t, "标量目标创建 SingleScanIterator",
		Expect(err, Equal(error(nil))),
		Expect(singleIter.(*SingleScanIterator).MustHasRecord(), Equal(false)),
		ExpectDo(func() error { return singleIter.Next(&target) }),
	)
	Then(t, "SingleScanIterator 记录已有结果",
		Expect(singleIter.(*SingleScanIterator).MustHasRecord(), Equal(true)),
	)
}

func TestRecvFuncItems(t *testing.T) {
	recv := RecvFunc[int](func(ctx context.Context, recv func(v *int) error) error {
		for _, n := range []int{1, 2} {
			v := n
			if err := recv(&v); err != nil {
				return err
			}
		}
		return nil
	})

	collected := make([]int, 0)
	for item, err := range recv.Items(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		collected = append(collected, *item)
	}

	Then(t, "RecvFunc.Items 逐项产出结果",
		Expect(collected, Equal([]int{1, 2})),
	)

	firstOnly := make([]int, 0)
	for item := range slices.Values(slices.Collect(func(yield func(*int) bool) {
		for item, err := range recv.Items(context.Background()) {
			if err != nil {
				t.Fatal(err)
			}
			if !yield(item) {
				return
			}
		}
	})) {
		firstOnly = append(firstOnly, *item)
		break
	}

	Then(t, "停止消费时可提前结束遍历",
		Expect(firstOnly, Equal([]int{1})),
	)
}
