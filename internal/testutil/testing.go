package testutil

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func Expect[A any](t testing.TB, actual A, matcheres ...testingx.Matcher[A]) {
	t.Helper()
	testingx.Expect[A](t, actual, matcheres...)
}

func Not[A any](m testingx.Matcher[A]) testingx.Matcher[A] {
	return testingx.Not(m)
}

func Be[A any](e A) testingx.Matcher[A] {
	return testingx.Be(e)
}

func Equal[A any](e A) testingx.Matcher[A] {
	return testingx.Equal(e)
}

func HaveLen[A any](c int) testingx.Matcher[A] {
	return testingx.HaveLen[A](c)
}
