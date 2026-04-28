package xiter

import (
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestSeqOperators(t *testing.T) {
	values := Of(1, 2, 3, 4)

	Then(t, "Map 和 Filter 保持 iter.Seq 的惰性组合语义",
		Expect(
			slices.Collect(Map(Filter(values, func(v int) bool {
				return v%2 == 0
			}), func(v int) string {
				return string(rune('0' + v))
			})),
			Equal([]string{"2", "4"}),
		),
	)

	visited := make([]int, 0)
	for v := range Of(1, 2, 3) {
		visited = append(visited, v)
		break
	}

	Then(t, "yield 返回 false 时 Of 停止遍历",
		Expect(visited, Equal([]int{1})),
	)
}
