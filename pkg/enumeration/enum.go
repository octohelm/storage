package enumeration

type CanEnumValues interface {
	EnumValues() []any
}

// DriverValueOffset
// sql value maybe have offset from const value in go
type DriverValueOffset interface {
	Offset() int
}
