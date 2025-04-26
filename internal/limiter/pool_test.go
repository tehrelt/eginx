package limiter_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tehrelt/eginx/internal/limiter"
	"github.com/tehrelt/eginx/internal/limiter/tokenbucket"
)

type mockstorage struct {
	db map[string]int
	m  sync.Mutex
}

func (s *mockstorage) Get(_ context.Context, key string) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if rate, ok := s.db[key]; ok {
		return rate, nil
	}

	return 0, errors.New("not found")
}

func TestPool(t *testing.T) {
	ctx := t.Context()
	key := "test"
	key1req := "one"
	storage := &mockstorage{db: map[string]int{
		key:     10,
		key1req: 1,
	}}

	create := func(ctx context.Context, rate int) limiter.Limiter {
		return tokenbucket.New(ctx, rate)
	}

	cases := []struct {
		name           string
		key            string
		requests       int
		wait_replenish bool
		allowed        bool
	}{
		{
			name:           "allowed",
			key:            key,
			requests:       10,
			allowed:        true,
			wait_replenish: false,
		},
		{
			name:           "forbidden",
			key:            key1req,
			requests:       2,
			allowed:        false,
			wait_replenish: false,
		},
		{
			name:           "allow_after_replenish",
			key:            key1req,
			requests:       2,
			allowed:        true,
			wait_replenish: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := limiter.NewLimiterPool(ctx, storage, create)
			result := false
			for range tc.requests {
				result = pool.Allow(t.Context(), tc.key)
				if !result && tc.wait_replenish {
					t.Log("waiting replenish")
					time.Sleep(time.Second + (10 * time.Millisecond))
					result = pool.Allow(t.Context(), tc.key)
				}
			}
			require.Equal(t, tc.allowed, result)
		})
	}

}
