//go:generate go tool gen .
package filter

import (
	"github.com/octohelm/storage/testdata/model"
)

// +gengo:filter
type filterOf struct {
	model.User
	model.Org
}
