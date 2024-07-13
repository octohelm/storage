package directive

import (
	"github.com/davecgh/go-spew/spew"
	testingx "github.com/octohelm/x/testing"
	"testing"
)

func Eq[T comparable](v T) Directive {
	return Directive{
		Name: "eq",
		Args: []any{v},
	}
}

func TestEncoder(t *testing.T) {
	x, _ := MarshalDirective("fn", 1, 2, 3, 4)
	testingx.Expect(t, string(x), testingx.Be(`fn(1,2,3,4)`))

	x2, err := MarshalDirective("or", Eq(1), Eq(2))
	if err != nil {
		spew.Dump(err)
	}
	testingx.Expect(t, string(x2), testingx.Be(`or(eq(1),eq(2))`))
}
