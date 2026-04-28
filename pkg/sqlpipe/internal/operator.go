package internal

// Operator 表示从输入到输出的泛型操作符。
type Operator[I any, O any] interface {
	Next(i I) O
}
