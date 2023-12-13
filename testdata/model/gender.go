package model

// +gengo:enum
type Gender int

const (
	GENDER_UNKNOWN Gender = iota
	GENDER__MALE          // 男
	GENDER__FEMALE        // 女
)

// +gengo:enum
type GenderExt string

const (
	GENDER_EXT__MALE   GenderExt = "MAIL"   // 男
	GENDER_EXT__FEMALE GenderExt = "FEMALE" // 女
)
