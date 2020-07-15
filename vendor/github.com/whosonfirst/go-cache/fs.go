package cache

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	_ "log"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type FSCache struct {
	Cache
	root           string
	misses         int64
	hits           int64
	evictions      int64
	mu             *sync.RWMutex
	TTL            int64
	FilePerms      os.FileMode
	DirectoryPerms os.FileMode
}

func init() {
	ctx := context.Background()
	RegisterCache(ctx, "fs", NewFSCache)
}

func NewFSCache(ctx context.Context, uri string) (Cache, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	root := u.Path

	abs_root, err := filepath.Abs(root)

	if err != nil {
		return nil, err
	}

	info, err := os.Stat(abs_root)

	if os.IsNotExist(err) {
		return nil, errors.New("Root doesn't exist")
	}

	if !info.IsDir() {
		return nil, errors.New("Root is not a directory")
	}

	mu := new(sync.RWMutex)

	c := &FSCache{
		hits:           int64(0),
		misses:         int64(0),
		mu:             mu,
		TTL:            0,
		FilePerms:      0644,
		DirectoryPerms: 0755,
		root:           root,
	}

	return c, nil
}

func (c *FSCache) Close(ctx context.Context) error {
	return nil
}

func (c *FSCache) Name() string {
	return "fs"
}

func (c *FSCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {

	c.mu.RLock()
	defer c.mu.RUnlock()

	abs_path := filepath.Join(c.root, key)

	info, err := os.Stat(abs_path)

	if err != nil {

		atomic.AddInt64(&c.misses, 1)

		if os.IsNotExist(err) {
			return nil, new(CacheMiss)
		}

		return nil, err
	}

	if c.TTL > 0 {

		mtime := info.ModTime()
		now := time.Now()

		diff := now.Sub(mtime)

		if diff >= time.Duration(c.TTL)*time.Second {

			go c.Unset(ctx, key)

			return nil, new(CacheMiss)
		}
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&c.hits, 1)
	return fh, nil
}

func (c *FSCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	abs_path := filepath.Join(c.root, key)
	abs_root := filepath.Dir(abs_path)

	_, err := os.Stat(abs_root)

	if os.IsNotExist(err) {

		err = os.MkdirAll(abs_root, c.DirectoryPerms)

		if err != nil {
			return nil, err
		}
	}

	out, err := os.OpenFile(abs_path, os.O_RDWR|os.O_CREATE, c.FilePerms)

	if err != nil {

		return nil, err
	}

	// would that we could do this but it always results in the following error:
	// 2018/06/17 08:28:30 write /usr/local/whosonfirst/go-whosonfirst-cache/foo: file already closed

	// defer out.Close()
	// r := io.TeeReader(fh, out)
	// return ioutil.NopCloser(r), nil

	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	wr := io.MultiWriter(out, buf)

	_, err = io.Copy(wr, fh)

	out.Close()
	buf.Flush()

	if err != nil {
		return nil, err
	}

	return NewReadCloser(b.Bytes()), nil
}

func (c *FSCache) Unset(ctx context.Context, key string) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	abs_path := filepath.Join(c.root, key)
	abs_root := filepath.Dir(abs_path)

	_, err := os.Stat(abs_root)

	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(abs_path)

	if err == nil {
		atomic.AddInt64(&c.evictions, 1)
	}

	return err
}

func (c *FSCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *FSCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *FSCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}

func (c *FSCache) Size() int64 {
	return c.SizeWithContext(context.Background())
}

func (c *FSCache) SizeWithContext(ctx context.Context) int64 {
	// TO DO: walk c.root
	return -1
}
