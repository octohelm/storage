package nullable

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"
)

func BenchmarkNewNullIgnoreScanner(b *testing.B) {
	v := 0
	for i := 0; i < b.N; i++ {
		_ = NewNullIgnoreScanner(&v).Scan(2)
	}
	b.Log(v)
}

func TestNullIgnoreScanner(t *testing.T) {
	t.Run("scan value", func(t *testing.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		_ = s.Scan(2)

		testutil.Expect(t, v, testutil.Equal(2))
	})

	t.Run("scan nil", func(t *testing.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		_ = s.Scan(nil)

		testutil.Expect(t, v, testutil.Equal(0))
	})
}
