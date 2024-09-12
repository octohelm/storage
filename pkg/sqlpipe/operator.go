package sqlpipe

type OperatorType uint

const (
	OperatorFilter OperatorType = iota + 1
	OperatorSort
	OperatorLimit
	OperatorOnConflict
	OperatorProject
	OperatorCommit
	OperatorSetting
)
