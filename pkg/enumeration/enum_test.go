package enumeration_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/enumeration"
)

func TestScanEnum(t *testing.T) {
	cases := []struct {
		offset int
		values []interface{}
		expect []int
	}{
		{
			-3,
			[]interface{}{
				nil,
				[]byte("-3"),
				"-2",
				int(-1),
				int8(0),
				int16(1),
				int32(2),
				int64(3),
				uint(4),
				uint8(5),
				uint16(6),
				uint32(7),
				uint64(8),
			},
			[]int{
				0,
				0,
				1,
				2,
				3,
				4,
				5,
				6,
				7,
				8,
				9,
				10,
				11,
			},
		},
	}

	for _, c := range cases {
		for i, v := range c.values {
			n, err := enumeration.ScanIntEnumStringer(v, c.offset)
			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, n, testutil.Equal(c.expect[i]))
		}
	}
}
