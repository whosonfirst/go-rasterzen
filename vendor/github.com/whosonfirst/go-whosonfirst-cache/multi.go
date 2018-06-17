package cache

import (
       "fmt"
	"io"
	"log"
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

func (mc *MultiCache) Get(key string) (io.ReadCloser, error) {

	var fh io.ReadCloser
	var err error

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	missing := false

	for _, c := range mc.caches {

		fh, err = c.Get(key)

		log.Printf("GET %T %s %V\n", c, key, err)

		if err != nil {

			if IsCacheMiss(err){
				missing = true
			}

			continue
		}

		break
	}

	if missing {
		err = new(CacheMissMulti)
	}

	return fh, err
}

func (mc *MultiCache) Set(key string, fh io.ReadCloser) (io.ReadCloser, error) {

     	log.Println("SET", key)

	var in io.ReadCloser
	var out io.ReadCloser
	var err error

	out = fh

	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, c := range mc.caches {

		in = out
		out, err = c.Set(key, in)

		log.Printf("SET %T %s %V\n", c, key, err)

		if err != nil {

			go mc.Unset(key)
			return nil, err
		}
	}

	return out, nil
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
