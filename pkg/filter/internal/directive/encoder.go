package directive

import (
	"bytes"
	"encoding/json"
	"strings"
)

func MarshalDirective(funcName string, args ...any) ([]byte, error) {
	b := bytes.NewBuffer(nil)

	b.WriteString(strings.ToLower(funcName))

	b.WriteByte('(')

	for i, arg := range args {
		if i > 0 {
			b.WriteByte(',')
		}

		switch x := arg.(type) {
		case Marshaler:
			directive, err := x.MarshalDirective()
			if err != nil {
				return nil, err
			}
			b.Write(directive)
		default:
			value, err := json.Marshal(arg)
			if err != nil {
				return nil, err
			}
			b.Write(value)
		}
	}

	b.WriteByte(')')

	return b.Bytes(), nil
}
