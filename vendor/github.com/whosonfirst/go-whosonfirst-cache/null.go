package cache

import (
	"io"
	"sync/atomic"
)

type NullCache struct {
	Cache
	misses int64
}

func NewNullCache() (Cache, error) {

	lc := NullCache{
		misses: int64(0),
	}

	return &lc, nil
}

func (c *NullCache) Name() string {
	return "null"
}

func (c *NullCache) Get(key string) (io.ReadCloser, error) {
	atomic.AddInt64(&c.misses, 1)
	return nil, new(CacheMiss)
}

func (c *NullCache) Set(key string, fh io.ReadCloser) (io.ReadCloser, error) {
	return fh, nil
}

func (c *NullCache) Unset(key string) error {
	return nil
}

func (c *NullCache) Size() int64 {
	return 0
}

func (c *NullCache) Hits() int64 {
	return 0
}

func (c *NullCache) Misses() int64 {
	return c.misses
}

func (c *NullCache) Evictions() int64 {
	return 0
}
