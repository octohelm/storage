package model

// +gengo:enum
type Gender int

const (
	GENDER_UNKNOWN Gender = iota
	GENDER__MALE
	GENDER__FEMALE
)
