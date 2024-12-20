package sqlfrag

import (
	"crypto/sha1"
	"fmt"
	"github.com/octohelm/x/slices"
	"strings"
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
	v, _ := a.m.LoadOrStore(name, sync.OnceValue(func() string {
		if len(name) < 32 {
			return name
		}
		hashed := fmt.Sprintf("%x", sha1.Sum([]byte(name)))

		return strings.Join(slices.Map(strings.Split(name, "_"), func(p string) string {
			if p == "t" {
				return "t_"
			}
			if len(p) > 0 {
				return p[0:1]
			}
			return p
		}), "") + "_" + hashed[0:8]
	}))
	return v.(func() string)()
}
