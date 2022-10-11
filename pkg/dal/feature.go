package dal

func IncludeAllRecord() func(m featureSettings) {
	return func(m featureSettings) {
		m.SetSoftDelete(false)
	}
}

func HardDelete() func(m featureSettings) {
	return func(m featureSettings) {
		m.SetSoftDelete(false)
	}
}

func WhereStmtNotEmpty() func(m featureSettings) {
	return func(m featureSettings) {
		m.SetWhereStmtNotEmpty(true)
	}
}

type OptionFunc func(m featureSettings)

type featureSettings interface {
	SetSoftDelete(flag bool)
	SetWhereStmtNotEmpty(flag bool)
}

type feature struct {
	softDelete        bool
	whereStmtNotEmpty bool
}

func (f *feature) SetSoftDelete(softDelete bool) {
	f.softDelete = softDelete
}

func (f *feature) SetWhereStmtNotEmpty(whereStmtNotEmpty bool) {
	f.whereStmtNotEmpty = whereStmtNotEmpty
}
