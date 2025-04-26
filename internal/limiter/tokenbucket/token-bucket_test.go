package tokenbucket_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tehrelt/eginx/internal/limiter/tokenbucket"
)

func TestAllowed(t *testing.T) {
	ctx := t.Context()
	limiter := tokenbucket.NewTokenBucket(ctx, 10, time.Second)

	for i := 0; i < 10; i++ {
		allowed := limiter.Allow()
		require.Equal(t, true, allowed)
	}
}

func TestForbidden(t *testing.T) {
	ctx := t.Context()
	limiter := tokenbucket.NewTokenBucket(ctx, 1, time.Second)

	allowed := limiter.Allow()
	require.Equal(t, true, allowed)
	allowed = limiter.Allow()
	require.Equal(t, false, allowed)
}

func TestAllowAfterForbid(t *testing.T) {
	ctx := t.Context()
	l := tokenbucket.NewTokenBucket(ctx, 1, time.Second)

	allowed := l.Allow()
	require.Equal(t, true, allowed)

	allowed = l.Allow()
	require.Equal(t, false, allowed)
	time.Sleep(time.Second)

	allowed = l.Allow()
	require.Equal(t, true, allowed)
}
