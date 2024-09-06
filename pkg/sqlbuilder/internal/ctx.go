package internal

import (
	contextx "github.com/octohelm/x/context"
)

var TableNameContext = contextx.New[string]()

var TableAliasContext = contextx.New[string]()
