package sqlfrag

import (
	"crypto/sha1"
	"fmt"
	"sync"
)

var a = &aliases{}

func SafeProjected(tableName string, colName string) string {
	// alias name max length is 63
	if len(tableName)+len(colName) < 64-2 {
		return fmt.Sprintf("%s__%s", tableName, colName)
	}
	return fmt.Sprintf("%s__%s", a.Safe(tableName), a.Safe(colName))
}

type aliases struct {
	m sync.Map
}

func (a *aliases) Safe(name string) string {
	v, _ := a.m.LoadOrStore(name, func() string {
		if len(name) < 32 {
			return name
		}
		hashed := fmt.Sprintf("%x", sha1.Sum([]byte(name)))
		return name[0:16] + hashed[0:8]
	})
	return v.(func() string)()
}
