//go:generate go tool gen .
package filter

import (
	"github.com/octohelm/storage/testdata/model"
)

// +gengo:filterop
type filterOf struct {
	model.User
	model.Org
	model.OrgUser
}
