package session

import (
	"github.com/octohelm/storage/internal/sql/scanner"
)

// Scan 复用底层 scanner.Scan。
var Scan = scanner.Scan

// ScanIterator 复用底层扫描迭代器接口。
type ScanIterator = scanner.ScanIterator

// Recv 将类型化回调包装为会话扫描迭代器。
func Recv[T any](next func(v *T) error) ScanIterator {
	return scanner.Recv(next)
}
