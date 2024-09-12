package internal

type Operator[I any, O any] interface {
	Next(i I) O
}
