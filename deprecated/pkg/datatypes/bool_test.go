package datatypes

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/storage/internal/testutil"
)

func TestBool(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		bytes, _ := json.Marshal(BOOL_TRUE)
		testutil.Expect(t, string(bytes), testutil.Equal("true"))

		bytes, _ = json.Marshal(BOOL_FALSE)
		testutil.Expect(t, string(bytes), testutil.Equal("false"))

		bytes, _ = json.Marshal(BOOL_UNKNOWN)
		testutil.Expect(t, string(bytes), testutil.Equal("null"))
	})
	t.Run("Unmarshal", func(t *testing.T) {
		var b Bool

		_ = json.Unmarshal([]byte("null"), &b)
		testutil.Expect(t, b, testutil.Equal(BOOL_UNKNOWN))

		_ = json.Unmarshal([]byte("true"), &b)
		testutil.Expect(t, b, testutil.Equal(BOOL_TRUE))

		_ = json.Unmarshal([]byte("false"), &b)
		testutil.Expect(t, b, testutil.Equal(BOOL_FALSE))
	})
}
