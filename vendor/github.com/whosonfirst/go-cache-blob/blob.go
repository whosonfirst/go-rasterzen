package cache

import (
	"bufio"
	"bytes"
	"context"
	wof_cache "github.com/whosonfirst/go-cache"
	"gocloud.dev/blob"
	"io"
	"io/ioutil"
	"sync/atomic"
)

type BlobCacheOptionsKey string

type BlobCache struct {
	wof_cache.Cache
	TTL       int64
	bucket    *blob.Bucket
	hits      int64
	misses    int64
	sets      int64
	evictions int64
}

func init() {
	ctx := context.Background()
	for _, scheme := range blob.DefaultURLMux().BucketSchemes() {
		wof_cache.RegisterCache(ctx, scheme, NewBlobCache)
	}
}

func NewBlobCache(ctx context.Context, uri string) (wof_cache.Cache, error) {

	bucket, err := blob.OpenBucket(ctx, uri)

	if err != nil {
		return nil, err
	}

	c := &BlobCache{
		TTL:       0,
		misses:    0,
		sets:      0,
		evictions: 0,
		bucket:    bucket,
	}

	return c, nil
}

func (c *BlobCache) Close(ctx context.Context) error {
	return nil
}

func (c *BlobCache) Name() string {
	return "blob"
}

func (c *BlobCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {

	fh, err := c.bucket.NewReader(ctx, key, nil)

	if err != nil {
		atomic.AddInt64(&c.misses, 1)
		return nil, err
	}

	atomic.AddInt64(&c.hits, 1)
	return fh, nil
}

func (c *BlobCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {

	var wr_opts *blob.WriterOptions

	v := ctx.Value(BlobCacheOptionsKey("options"))

	if v != nil {
		wr_opts = v.(*blob.WriterOptions)
	}
	
	bucket_wr, err := c.bucket.NewWriter(ctx, key, wr_opts)

	if err != nil {
		return nil, err
	}

	// this is not awesome but until we update all the things (and
	// in particular all the go-whosonfirst-readwrite stuff) to be
	// ReadSeekCloser thingies it's what necessary...
	// (20180617/thisisaaronland)

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	io.Copy(wr, fh)
	wr.Flush()

	r := bytes.NewReader(b.Bytes())

	_, err = io.Copy(bucket_wr, r)

	if err != nil {
		return nil, err
	}

	err = bucket_wr.Close()

	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&c.sets, 1)

	r.Reset(b.Bytes())
	return ioutil.NopCloser(r), nil
}

func (c *BlobCache) Unset(ctx context.Context, key string) error {

	err := c.bucket.Delete(ctx, key)

	if err != nil {
		return err
	}

	atomic.AddInt64(&c.evictions, 1)
	return nil
}

func (c *BlobCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *BlobCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *BlobCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}

func (c *BlobCache) Size() int64 {

	return c.SizeWithContext(context.Background())
}

func (c *BlobCache) SizeWithContext(ctx context.Context) int64 {

	size := int64(0)

	iter := c.bucket.List(nil)

	for {

		select {
		case <-ctx.Done():
			return -1
		default:
			//
		}

		obj, err := iter.Next(ctx)

		if err == io.EOF {
			break
		}

		if err != nil {
			return -1
		}

		size += obj.Size
	}

	return size
}
