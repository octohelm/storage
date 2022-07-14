package model

// +gengo:enum
type Gender int

const (
	GENDER_UNKNOWN Gender = iota
	GENDER__MALE          // 男
	GENDER__FEMALE        // 女
)
