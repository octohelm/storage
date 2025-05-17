package sqlpipe

type OperatorType uint

const (
	OperatorCompose OperatorType = iota
	OperatorProject
	OperatorFilter
	OperatorSort
	OperatorLimit
	OperatorLock
	OperatorOnConflict
	OperatorJoin
	OperatorCommit
	OperatorSetting
)
