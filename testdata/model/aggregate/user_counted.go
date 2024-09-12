package aggregate

import "github.com/octohelm/storage/testdata/model"

// +gengo:table
type CountedUser struct {
	model.User

	Count int `db:"f_count"`
}
