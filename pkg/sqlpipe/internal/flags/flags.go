package flags

const (
	none = -(iota + 1)
	whereRequired
	includesAll
	forReturning

	withoutSorter
	withoutPager
)

const (
	WhereRequired Flag = 1 << -whereRequired
	IncludesAll   Flag = 1 << -includesAll
	ForReturning  Flag = 1 << -forReturning

	WithoutSorter Flag = 1 << -withoutSorter
	WithoutPager  Flag = 1 << -withoutPager
)

type Flag uint64

func (f Flag) Is(mode Flag) bool {
	return f&mode != 0
}

func (f Flag) With(f2 Flag) Flag {
	return f | f2
}
