package time

import (
	"testing"
	"time"

	"github.com/octohelm/storage/internal/testutil"
)

func TestTimestamp(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Timestamp(t0)

		testutil.Expect(t, dt.String(), testutil.Equal("2017-03-27T23:58:59+08:00"))
		testutil.Expect(t, dt.Format(time.RFC3339), testutil.Equal("2017-03-27T23:58:59+08:00"))
		testutil.Expect(t, dt.Unix(), testutil.Equal(int64(1490630339)))
	})
	t.Run("Marshal & Unmarshal", func(t *testing.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Timestamp(t0)

		dateString, err := dt.MarshalText()
		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, string(dateString), testutil.Equal("2017-03-27T23:58:59+08:00"))

		dt2 := TimestampZero
		testutil.Expect(t, dt2.IsZero(), testutil.Be(true))

		err = dt2.UnmarshalText(dateString)
		testutil.Expect(t, err, testutil.Be[error](nil))
		testutil.Expect(t, dt2, testutil.Equal(dt))
		testutil.Expect(t, dt2.IsZero(), testutil.Be(false))

		dt3 := TimestampZero
		err = dt3.UnmarshalText([]byte(""))
		testutil.Expect(t, err, testutil.Be[error](nil))
	})
}
