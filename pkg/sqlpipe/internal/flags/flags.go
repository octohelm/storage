package flags

const (
	none = -(iota + 1)
	whereOrLimitRequired
	includesAll
	forReturning

	withoutSorter
	withoutPager
)

const (
	WhereOrPagerRequired Flag = 1 << -whereOrLimitRequired
	IncludesAll          Flag = 1 << -includesAll
	ForReturning         Flag = 1 << -forReturning

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

func (f Flag) Without(f2 Flag) Flag {
	return f &^ f2
}
