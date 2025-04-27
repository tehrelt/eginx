package tokenbucket_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tehrelt/eginx/internal/limiter/tokenbucket"
)

func TestTokenBucketPool(t *testing.T) {
	storage := &mockdb{
		db: map[string]int{
			"key1":      10,
			"key2":      20,
			"replenish": 60,
		},
	}

	cases := []struct {
		name                 string
		keys                 []string
		requests             int
		wait                 time.Duration
		lastRequestIsAllowed bool
	}{
		{
			name:                 "all allow",
			keys:                 []string{"key1", "key2"},
			requests:             10,
			lastRequestIsAllowed: true,
		},
		{
			name:                 "all forbid",
			keys:                 []string{"key1", "key2"},
			requests:             100,
			lastRequestIsAllowed: false,
		},
		{
			name:                 "wait replenish to allow",
			keys:                 []string{"replenish"},
			requests:             61,
			wait:                 time.Second,
			lastRequestIsAllowed: true,
		},
	}

	t.Parallel()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := tokenbucket.NewTokenBucket(t.Context(), storage)

			for _, key := range tc.keys {
				result := false
				for i := 0; i < tc.requests; i++ {
					result = pool.Allow(t.Context(), key)
					if !result && tc.wait > 0 {
						time.Sleep(tc.wait)
						result = pool.Allow(t.Context(), key)
					}
				}

				require.Equal(t, tc.lastRequestIsAllowed, result)
			}
		})
	}
}

type mockdb struct {
	db map[string]int
	m  sync.Mutex
}

func (m *mockdb) Get(_ context.Context, key string) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()
	return m.db[key], nil
}
