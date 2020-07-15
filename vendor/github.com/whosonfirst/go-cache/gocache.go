package cache

// https://godoc.org/github.com/patrickmn/go-cache

import (
	"context"
	gocache "github.com/patrickmn/go-cache"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

func init() {
	ctx := context.Background()
	RegisterCache(ctx, "gocache", NewGoCache)
}

type GoCacheOptions struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

type GoCache struct {
	Cache
	cache     *gocache.Cache
	hits      int64
	misses    int64
	evictions int64
	keys      int64
}

func DefaultGoCacheOptions() (*GoCacheOptions, error) {

	opts := GoCacheOptions{
		DefaultExpiration: 0 * time.Second,
		CleanupInterval:   0 * time.Second,
	}

	return &opts, nil
}

func NewGoCache(ctx context.Context, uri string) (Cache, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	opts, err := DefaultGoCacheOptions()

	// TODO: parse out 1m, etc. strings
	// (20171229/thisisaaronland)

	str_exp := q.Get("default_expiration")

	if str_exp != "" {

		exp, err := strconv.Atoi(str_exp)

		if err != nil {
			return nil, err
		}

		opts.DefaultExpiration = time.Duration(exp) * time.Second
	}

	str_cleanup := q.Get("cleanup_interval")

	if str_cleanup != "" {

		cleanup, err := strconv.Atoi(str_cleanup)

		if err != nil {
			return nil, err
		}

		opts.CleanupInterval = time.Duration(cleanup) * time.Second
	}

	gc := gocache.New(opts.DefaultExpiration, opts.CleanupInterval)

	c := &GoCache{
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
		keys:      0,
		cache:     gc,
	}

	return c, nil
}

func (c *GoCache) Close(ctx context.Context) error {
	return nil
}

func (c *GoCache) Name() string {
	return "gocache"
}

func (c *GoCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	data, ok := c.cache.Get(key)

	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return nil, new(CacheMiss)
	}

	atomic.AddInt64(&c.hits, 1)

	body := data.([]byte)

	return NewReadCloser(body), nil
}

func (c *GoCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {

	/*

	   Assume an io.Reader is hooked up to a satellite dish receiving a message (maybe a 1TB message) from an
	   alien civilization who only transmits their message once every thousand years. There's no "rewinding"
	   that.

	   https://groups.google.com/forum/#!msg/golang-nuts/BzDAg0CFqyk/t3TvH9QV0xEJ

	*/

	defer fh.Close()

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	c.cache.Set(key, body, gocache.DefaultExpiration)
	atomic.AddInt64(&c.keys, 1)

	return NewReadCloser(body), nil
}

func (c *GoCache) Unset(ctx context.Context, key string) error {
	c.cache.Delete(key)
	atomic.AddInt64(&c.keys, -1)
	return nil
}

func (c *GoCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *GoCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *GoCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}

func (c *GoCache) Size() int64 {
	return c.SizeWithContext(context.Background())
}

func (c *GoCache) SizeWithContext(ctx context.Context) int64 {
	return atomic.LoadInt64(&c.keys)
}
