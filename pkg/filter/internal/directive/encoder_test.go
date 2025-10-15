package directive

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
	"github.com/octohelm/x/testing/bdd"
)

func Eq[T comparable](v T) Directive {
	return Directive{
		Name: "eq",
		Args: []any{v},
	}
}

func TestEncoder(t *testing.T) {
	x := bdd.Must(MarshalDirective("fn", 1, 2, 3, 4))
	testingx.Expect(t, string(x), testingx.Be(`fn(1,2,3,4)`))

	x2 := bdd.Must(MarshalDirective("or", Eq(1), Eq(2)))
	testingx.Expect(t, string(x2), testingx.Be(`or(eq(1),eq(2))`))
}
