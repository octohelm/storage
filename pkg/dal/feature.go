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

type OptionFunc func(m featureSettings)

type featureSettings interface {
	SetSoftDelete(flag bool)
}

type feature struct {
	softDelete bool
}

func (f *feature) SetSoftDelete(softDelete bool) {
	f.softDelete = softDelete
}
