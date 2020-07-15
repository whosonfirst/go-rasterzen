package cache

import (
	"context"
	"io"
	"sync/atomic"
)

type NullCache struct {
	Cache
	misses int64
}

func init() {
	ctx := context.Background()
	RegisterCache(ctx, "null", NewNullCache)
}

func NewNullCache(ctx context.Context, uri string) (Cache, error) {
	c := &NullCache{
		misses: int64(0),
	}
	return c, nil
}

func (c *NullCache) Close(ctx context.Context) error {
	return nil
}

func (c *NullCache) Name() string {
	return "null"
}

func (c *NullCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	atomic.AddInt64(&c.misses, 1)
	return nil, new(CacheMiss)
}

func (c *NullCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {
	return fh, nil
}

func (c *NullCache) Unset(ctx context.Context, key string) error {
	return nil
}

func (c *NullCache) Size() int64 {
	return 0
}

func (c *NullCache) SizeWithContext(ctx context.Context) int64 {
	return 0
}

func (c *NullCache) Hits() int64 {
	return 0
}

func (c *NullCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *NullCache) Evictions() int64 {
	return 0
}
