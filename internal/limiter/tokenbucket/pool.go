package tokenbucket

import (
	"context"
	"errors"
	"sync"

	"github.com/tehrelt/eginx/internal/storage"
)

type KeyStorage interface {
	Get(context.Context, string) (int, error)
}

type bucketStorage struct {
	db map[string]*bucket
	m  sync.Mutex
}

func newBucketStorage() bucketStorage {
	return bucketStorage{
		db: make(map[string]*bucket),
	}
}

func (bs *bucketStorage) get(key string) (*bucket, bool) {
	bs.m.Lock()
	defer bs.m.Unlock()

	b, ok := bs.db[key]
	return b, ok
}
func (bs *bucketStorage) set(key string, b *bucket) {
	bs.m.Lock()
	defer bs.m.Unlock()

	bs.db[key] = b
}

type TokenBucket struct {
	mainCtx       context.Context
	buckets       bucketStorage
	storage       KeyStorage
	defaultBucket *bucket
}

type TokenBucketOpt func(*TokenBucket)

func WithDefaultBucket(ctx context.Context, rate int) TokenBucketOpt {
	return func(tb *TokenBucket) {
		tb.defaultBucket = NewBucket(ctx, rate)
	}
}

func NewTokenBucket(ctx context.Context, storage KeyStorage, opts ...TokenBucketOpt) *TokenBucket {
	tb := &TokenBucket{
		mainCtx: ctx,
		buckets: newBucketStorage(),
		storage: storage,
	}

	for _, opt := range opts {
		opt(tb)
	}

	return tb
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) bool {
	b, err := tb.get(ctx, key)
	if err != nil {
		return false
	}

	return b.Allow()
}

func (tb *TokenBucket) get(ctx context.Context, key string) (*bucket, error) {
	b, ok := tb.buckets.get(key)
	if !ok {
		rate, err := tb.storage.Get(ctx, key)
		if err != nil {
			if errors.Is(err, storage.ErrClientNotFound) {
				if tb.defaultBucket == nil {
					return nil, storage.ErrClientNotFound
				}

				return tb.defaultBucket, nil
			}
		}

		b = tb.newBucket(rate)
		tb.buckets.set(key, b)
	} else {
		rate, err := tb.storage.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		if rate != b.rpm {
			b = tb.newBucket(rate)
			tb.buckets.set(key, b)
		}
	}

	return b, nil
}

func (tb *TokenBucket) newBucket(rate int) *bucket {
	return NewBucket(tb.mainCtx, rate)
}
