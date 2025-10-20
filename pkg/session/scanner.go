package session

import "github.com/octohelm/storage/internal/sql/scanner"

var Scan = scanner.Scan

type ScanIterator = scanner.ScanIterator

func Recv[T any](next func(v *T) error) ScanIterator {
	return scanner.Recv(next)
}
