package multicache

import (
	"bytes"
	"context"
	"github.com/whosonfirst/go-cache"
	"io"
	"io/ioutil"
	"sync/atomic"
)

type MultiCache struct {
	cache.Cache
	size      int64
	hits      int64
	misses    int64
	evictions int64
	caches    []cache.Cache
}

/*
func init() {
	ctx := context.Background()
	RegisterCache(ctx, "null", NewMultiCache)
}
*/

func NewMultiCache(ctx context.Context, caches ...cache.Cache) (cache.Cache, error) {
	c := &MultiCache{
		size:      int64(0),
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
		caches:    caches,
	}
	return c, nil
}

func (mc *MultiCache) Close(ctx context.Context) error {

	for _, c := range mc.caches {

		err := c.Close(ctx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (mc *MultiCache) Name() string {
	return "multi"
}

func (mc *MultiCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {

	for _, c := range mc.caches {

		fh, err := c.Get(ctx, key)

		if err != nil {
			continue
		}

		atomic.AddInt64(&mc.hits, 1)
		return fh, nil
	}

	atomic.AddInt64(&mc.misses, 1)
	return nil, new(cache.CacheMiss)
}

func (mc *MultiCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(body)

	for _, c := range mc.caches {

		br.Seek(0, 0)
		cl := ioutil.NopCloser(br)

		_, err := c.Set(ctx, key, cl)

		if err != nil {
			return nil, err
		}
	}

	atomic.AddInt64(&mc.size, 1)

	br.Seek(0, 0)
	cl := ioutil.NopCloser(br)

	return cl, nil
}

func (mc *MultiCache) Unset(ctx context.Context, key string) error {

	for _, c := range mc.caches {

		err := c.Unset(ctx, key)

		if err != nil {
			return err
		}
	}

	atomic.AddInt64(&mc.size, -1)
	atomic.AddInt64(&mc.evictions, 1)

	return nil
}

func (mc *MultiCache) Size() int64 {
	return mc.SizeWithContext(context.Background())
}

func (mc *MultiCache) SizeWithContext(ctx context.Context) int64 {
	return atomic.LoadInt64(&mc.size)
}

func (mc *MultiCache) Hits() int64 {
	return atomic.LoadInt64(&mc.hits)
}

func (mc *MultiCache) Misses() int64 {
	return atomic.LoadInt64(&mc.misses)
}

func (mc *MultiCache) Evictions() int64 {
	return atomic.LoadInt64(&mc.evictions)
}
