/*
Package compose GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package compose

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v List[T]) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Items":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (OneToMulti[ID, Record]) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (OneToOne[ID, Record]) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
