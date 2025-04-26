package iter

type Iterator[T any] interface {
	Next() (T, bool)
	Set([]T)
}
