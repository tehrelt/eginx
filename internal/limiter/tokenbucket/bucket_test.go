package tokenbucket_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tehrelt/eginx/internal/limiter/tokenbucket"
)

func TestAllowed(t *testing.T) {
	ctx := t.Context()
	bucket := tokenbucket.NewBucket(ctx, 10)

	for i := 0; i < 10; i++ {
		allowed := bucket.Allow()
		require.Equal(t, true, allowed)
	}
}

func TestForbidden(t *testing.T) {
	ctx := t.Context()
	bucket := tokenbucket.NewBucket(ctx, 1)

	allowed := bucket.Allow()
	require.Equal(t, true, allowed)
	allowed = bucket.Allow()
	require.Equal(t, false, allowed)
}

func TestAllowAfterForbid(t *testing.T) {
	ctx := t.Context()
	n := 120
	bucket := tokenbucket.NewBucket(ctx, n)

	actual := false
	expected := true

	for range n + 1 {
		actual = bucket.Allow()
	}
	require.Equal(t, false, actual)
	time.Sleep(time.Second)

	actual = bucket.Allow()
	require.Equal(t, expected, actual)
}
