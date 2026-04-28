package sqlpipe

// OperatorType 表示 sqlpipe 操作符的排序类型。
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
