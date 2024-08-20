/*
Package model GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package model

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (Gender) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (GenderExt) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v OperateTime) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "UpdatedAt":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v OperateTimeWithDeleted) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OperateTime":
			return []string{}, true
		case "DeletedAt":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.OperateTime, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Org) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ID":
			return []string{}, true
		case "Name":
			return []string{}, true
		case "OperateTimeWithDeleted":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.OperateTimeWithDeleted, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"Org",
	}, true
}

func (v OrgUser) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ID":
			return []string{}, true
		case "UserID":
			return []string{}, true
		case "OrgID":
			return []string{}, true

		}

		return nil, false
	}
	return []string{
		"OrgUser",
	}, true
}

func (v User) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ID":
			return []string{}, true
		case "Name":
			return []string{
				"姓名",
			}, true
		case "Nickname":
			return []string{}, true
		case "Username":
			return []string{}, true
		case "Gender":
			return []string{}, true
		case "Age":
			return []string{}, true
		case "OperateTimeWithDeleted":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.OperateTimeWithDeleted, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"User",
	}, true
}

func (v UserV2) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ID":
			return []string{}, true
		case "Nickname":
			return []string{}, true
		case "Gender":
			return []string{}, true
		case "Name":
			return []string{}, true
		case "RealName":
			return []string{}, true
		case "Age":
			return []string{
				"for modify testing",
			}, true
		case "Username":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}
