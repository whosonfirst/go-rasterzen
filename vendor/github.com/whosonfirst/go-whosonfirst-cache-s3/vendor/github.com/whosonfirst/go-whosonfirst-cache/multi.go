package cache

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	_ "log"
	"strings"
	"sync"
)

type MultiCache struct {
	Cache
	caches []Cache
	mu     *sync.RWMutex
}

type CacheMissMulti struct {
	error string
}

func (m CacheMissMulti) Error() string {

	return fmt.Sprintf("ONE OR MORE MULTI CACHE MISSES %s", m.error)
}

func IsCacheMissMulti(e error) bool {

	switch e.(type) {
	case *CacheMissMulti:
		return true
	case CacheMissMulti:
		return true
	default:
		// pass
	}

	return false
}

func NewMultiCache(caches []Cache) (Cache, error) {

	// test to make sure nothing is caches is a MultiCache...

	mu := new(sync.RWMutex)

	mc := MultiCache{
		caches: caches,
		mu:     mu,
	}

	return &mc, nil
}

func (mc *MultiCache) Name() string {

	cache_names := make([]string, len(mc.caches))

	for i, c := range mc.caches {
		cache_names[i] = c.Name()
	}

	return fmt.Sprintf("multi#%s", strings.Join(cache_names, ";"))
}

func (mc *MultiCache) Get(key string) (io.ReadCloser, error) {

	var fh io.ReadCloser
	var err error

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	missing := false

	for _, c := range mc.caches {

		fh, err = c.Get(key)

		if err != nil {

			if IsCacheMiss(err) {
				missing = true
			}

			continue
		}

		break
	}

	if missing {
		err = new(CacheMissMulti)
	}

	if fh == nil {
		err = new(CacheMiss)
	}

	return fh, err
}

// in advance of requiring a ReadSeekCloser (20180617/thisisaaronland)

func (mc *MultiCache) Set(key string, fh io.ReadCloser) (io.ReadCloser, error) {

	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	_, err := io.Copy(buf, fh)

	if err != nil {
		return nil, err
	}

	err = buf.Flush()

	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b.Bytes())

	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, c := range mc.caches {

		_, err := c.Set(key, ioutil.NopCloser(r))

		if err != nil {

			go mc.Unset(key)
			return nil, err
		}

		r.Reset(b.Bytes())
	}

	return ioutil.NopCloser(r), nil
}

func (mc *MultiCache) Unset(key string) error {

	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, c := range mc.caches {

		err := c.Unset(key)

		if err != nil {
			return err
		}
	}

	return nil
}

func (mc *MultiCache) Size() int64 {
	return 0
}

func (mc *MultiCache) Hits() int64 {
	return 0
}

func (mc *MultiCache) Misses() int64 {
	return 0
}

func (mc *MultiCache) Evictions() int64 {
	return 0
}
