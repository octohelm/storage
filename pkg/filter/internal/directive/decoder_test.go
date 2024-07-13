package directive

import (
	"bytes"
	testingx "github.com/octohelm/x/testing"
	"testing"
)

func TestNewDecoder(t *testing.T) {
	d := NewDecoder(bytes.NewBufferString(`fn("1",eq(1))`))
	d.RegisterDirectiveNewer(DefaultDirectiveNewer, func() Unmarshaler {
		return &Directive{}
	})

	fn := &Directive{}
	err := fn.UnmarshalDirective(d)
	testingx.Expect(t, err, testingx.BeNil[error]())

	v, _ := fn.MarshalDirective()
	testingx.Expect(t, string(v), testingx.Be(`fn("1",eq(1))`))
}
