/*
Package filter GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package filter

import _ "embed"

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

func (v *OrgByCreatedAt) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "CreatedAt":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *OrgByID) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ID":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *OrgByName) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *UserByAge) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Age":
			return []string{
				"年龄",
			}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *UserByCreatedAt) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "CreatedAt":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *UserByDeletedAt) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "DeletedAt":
			return []string{}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *UserByID) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ID":
			return []string{
				"用户ID",
			}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *UserByName) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{
				"姓名",
			}, true
		}

		return nil, false
	}
	return []string{}, true
}

func (v *UserByNickname) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Nickname":
			return []string{
				"昵称",
			}, true
		}

		return nil, false
	}
	return []string{}, true
}
