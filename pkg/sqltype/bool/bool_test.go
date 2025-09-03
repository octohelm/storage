package bool

import (
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestBool(t *testing.T) {
	b := bdd.FromT(t)

	b.Given("true", func(b bdd.T) {
		dt := BOOL_TRUE

		b.Then("marshal json",
			bdd.Equal("true", string(bdd.Must(dt.MarshalJSON()))),
		)

		var v Bool

		b.Then("unmarshal normal",
			bdd.NoError(v.UnmarshalJSON([]byte(`true`))),
			bdd.Equal(BOOL_TRUE, v),
		)

		b.Then("unmarshal with quote",
			bdd.NoError(v.UnmarshalJSON([]byte(`"false"`))),
			bdd.Equal(BOOL_FALSE, v),
		)
	})
}
