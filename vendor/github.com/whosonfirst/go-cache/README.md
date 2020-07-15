# go-cache

There are many interfaces for caching things. This one is ours. It reads and writes `io.ReadCloser` instances.

_This package supersedes [go-whosonfirst-cache](https://github.com/whosonfirst/go-whosonfirst-cache) which will be retired soon._

## Example

Caches are instantiated with the `cache.NewCache` method which takes as its arguments a `context.Context` instance and a URI string. The URI's scheme represents the type of cache it implements and the remaining (URI) properties are used by that cache type to instantiate itself.

For example to cache files to/from a directory on the local filesystem you would write:

```
import (
	"context"
	"flag"
	"github.com/whosonfirst/go-cache"
	"log"
)

func main() {

	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "fs:///usr/local/cache")

	str, err := cache.GetString(c, "some-key")

	if err != nil && !cache.IsCacheMiss(err) {
		log.Fatal(err)
	}

	str, _ := cache.SetString(c, "some-key", "some-value")

	str2, _ := cache.GetString(c, "some-key")

	log.Println(str2)
}
```

Two things to note:

* The use of the `fs://` scheme rather than the more conventional `file://`. This is deliberate so as not to overlap with the [Go Cloud](https://gocloud.dev/howto/blob/) `Blob` package's file handler.

* The use of the `cache.GetString` and `cache.SetString` methods. The `cache.Cache` interface expects `io.ReadCloser` instances so these methods are shortcuts to hide the boilerplate code necessary to work with `io.ReadCloser` interfaces.

There is also a handy `null://` cache which doesn't do anything at all (expect implement the `cache.Cache` interface). For example:

```
	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "null://")
```

## Interfaces

### cache.Cache

```
type Cache interface {
     	Name() string
	Close(context.Context) error
	Get(context.Context, string) (io.ReadCloser, error)
	Set(context.Context, string, io.ReadCloser) (io.ReadCloser, error)
	Unset(context.Context, string) error
	Hits() int64
	Misses() int64
	Evictions() int64
	Size() int64
	SizeWithContext(context.Context) int64
}
```

## Custom caches

Custom caches need to:

1. Implement the interface above.
2. Announce their availability using the `go-cache.Cache` method on initialization.

For example, here is an abbreviated example of how the [go-cache-blob](https://github.com/whosonfirst/go-cache-blob/) is implemented:

```
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
```

Note: See the way we're registering available cache types based on the the list of available Go Cloud `Bucket` schemes? By design the `go-cache-blob` doesn't try to load any Go Cloud providers. That is left up to your code.

```
func NewBlobCache() wof_cache.Cache {

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

	return c
}
```

And then to use it you would do this:

```
package main

import (
	"context"
	"github.com/whosonfirst/go-cache"
	_ "github.com/whosonfirst/go-cache-blob"
	_ "gocloud.dev/blob/memblob"
)

func main() {
	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "mem://")
	fh, _ := c.Get(ctx, "some-key")
}
```

## Available caches

### "blob"

Cache data using any registered [Go Cloud](https://gocloud.dev/howto/blob/) `Blob` source. For example:

```
import (
	"context"
	"github.com/whosonfirst/go-cache"
	_ "github.com/whosonfirst/go-cache-blob"
	_ "gocloud.dev/blob/s3blob"	
)

func main() {
	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "s3://cache-bucket?region=us-west-1")
}
```

* https://github.com/whosonfirst/go-cache-blob

### fs://

Cache data using a local filesystem.

```
import (
	"context"
	"github.com/whosonfirst/go-cache"
)

func main() {
	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "fs:///usr/local/cache")
}
```

* https://github.com/whosonfirst/go-cache

### gocache://

Cache data using a [go-cache](https://github.com/patrickmn/go-cache) backend.

```
import (
	"context"
	"github.com/whosonfirst/go-cache"
)

func main() {
	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "gocache://")
}
```

To specify custom values for the `DefaultExpiration` and `CleanupInterval` properties pass use the `default_expiration` and `cleanup_interval` parameters, respectively. For example:

```
	c, _ := cache.NewCache(ctx, "gocache://?default_expiration=300&cleanup_interval=200")
```

* https://github.com/whosonfirst/go-cache

### null://

Pretend to cache data.

```
import (
	"context"
	"github.com/whosonfirst/go-cache"
)

func main() {
	ctx := context.Background()
	c, _ := cache.NewCache(ctx, "null://")
}
```
