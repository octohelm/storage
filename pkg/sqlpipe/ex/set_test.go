package ex

import (
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestSets(t *testing.T) {
	oneToMulti := OneToMulti[int, string]{}
	a, b := "a", "b"
	oneToMulti.Record(1, &a)
	oneToMulti.Record(1, &b)

	Then(t, "OneToMulti 记录多个值",
		Expect(oneToMulti.IsZero(), Equal(false)),
		Expect(len(slices.Collect(oneToMulti.Keys())), Equal(1)),
		Expect(len(slices.Collect(oneToMulti.Records(1))), Equal(2)),
		Expect(len(slices.Collect(oneToMulti.AllRecords())), Equal(2)),
	)

	filled := make([]string, 0)
	oneToMulti.FillWith(1, func(p *string) { filled = append(filled, *p) })
	Then(t, "FillWith 遍历指定 key 的值",
		Expect(filled, Equal([]string{"a", "b"})),
	)

	oneToOne := OneToOne[int, string]{}
	oneToOne.Record(2, &a)
	Then(t, "OneToOne 只保留单值记录",
		Expect(oneToOne.IsZero(), Equal(false)),
		Expect(len(slices.Collect(oneToOne.Keys())), Equal(1)),
		Expect(len(slices.Collect(oneToOne.Records(2))), Equal(1)),
		Expect(len(slices.Collect(oneToOne.AllRecords())), Equal(1)),
	)

	all := slices.Collect(AllRecords[int, string](oneToMulti))
	first := slices.Collect(FirstRecords[int, string](oneToMulti))
	Then(t, "通用 Set 帮助函数返回全部值和首值",
		Expect(len(all), Equal(2)),
		Expect(len(first), Equal(1)),
		Expect(*first[0], Equal("a")),
	)
}
