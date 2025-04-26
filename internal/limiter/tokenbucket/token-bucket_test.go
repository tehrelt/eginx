package tokenbucket_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tehrelt/eginx/internal/limiter/tokenbucket"
)

func TestAllowed(t *testing.T) {
	ctx := t.Context()
	limiter := tokenbucket.New(ctx, 10)

	for i := 0; i < 10; i++ {
		allowed := limiter.Allow()
		require.Equal(t, true, allowed)
	}
}

func TestForbidden(t *testing.T) {
	ctx := t.Context()
	limiter := tokenbucket.New(ctx, 1)

	allowed := limiter.Allow()
	require.Equal(t, true, allowed)
	allowed = limiter.Allow()
	require.Equal(t, false, allowed)
}

func TestAllowAfterForbid(t *testing.T) {
	ctx := t.Context()
	l := tokenbucket.New(ctx, 1)

	allowed := l.Allow()
	require.Equal(t, true, allowed)

	allowed = l.Allow()
	require.Equal(t, false, allowed)
	time.Sleep(time.Second)

	allowed = l.Allow()
	require.Equal(t, true, allowed)
}
