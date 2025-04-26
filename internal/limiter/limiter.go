package limiter

type Limiter interface {
	Allow() bool
}

type LimitPolicy struct {
	Key        string
	RatePerSec int
}
