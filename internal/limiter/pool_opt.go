package limiter

func WithDefaultLimiter(limiter Limiter) LimiterPoolOpt {
	return func(lp *limiterPool) {
		lp.defaultLimiter = limiter
	}
}
