package testutil

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

// Expect 转发到底层测试断言。
func Expect[A any](t testing.TB, actual A, matcheres ...testingx.Matcher[A]) {
	t.Helper()
	testingx.Expect[A](t, actual, matcheres...)
}

// Not 返回取反匹配器。
func Not[A any](m testingx.Matcher[A]) testingx.Matcher[A] {
	return testingx.Not(m)
}

// Be 返回相等匹配器。
func Be[A any](e A) testingx.Matcher[A] {
	return testingx.Be(e)
}

// Equal 返回深度相等匹配器。
func Equal[A any](e A) testingx.Matcher[A] {
	return testingx.Equal(e)
}

// HaveLen 返回长度匹配器。
func HaveLen[A any](c int) testingx.Matcher[A] {
	return testingx.HaveLen[A](c)
}
