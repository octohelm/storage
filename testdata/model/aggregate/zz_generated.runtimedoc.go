/*
Package aggregate GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package aggregate

func (v *CountedUser) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Count":
			return []string{}, true
		}
		if doc, ok := runtimeDoc(&v.User, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

// nolint:deadcode,unused
func runtimeDoc(v any, prefix string, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		doc, ok := c.RuntimeDoc(names...)
		if ok {
			if prefix != "" && len(doc) > 0 {
				doc[0] = prefix + doc[0]
				return doc, true
			}

			return doc, true
		}
	}
	return nil, false
}
