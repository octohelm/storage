package internal

import (
	contextx "github.com/octohelm/x/context"
)

// TableNameContext 在上下文中传递当前表名。
var TableNameContext = contextx.New[string]()

// TableAliasContext 在上下文中传递当前表别名。
var TableAliasContext = contextx.New[string]()
