package limiter

type Limiter interface {
	Allow() bool
	Capacity() int
}

type LimitPolicy struct {
	Key        string
	RatePerSec int
}
