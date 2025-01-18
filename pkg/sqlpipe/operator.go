package sqlpipe

type OperatorType uint

const (
	OperatorCompose OperatorType = iota
	OperatorProject
	OperatorFilter
	OperatorSort
	OperatorLimit
	OperatorOnConflict
	OperatorJoin
	OperatorCommit
	OperatorSetting
)
