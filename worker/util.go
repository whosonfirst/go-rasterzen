package worker

import (
	"bufio"
	"bytes"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
)

func CheckCache(c cache.Cache, key string) (bool, error) {

	cache_data, err := c.Get(key)

	if err == nil {

		err = cache_data.Close()

		if err != nil {
			return true, err
		}

		return true, nil
	}

	if cache.IsCacheMissMulti(err) {

		var b bytes.Buffer
		wr := bufio.NewWriter(&b)

		_, err = io.Copy(wr, cache_data)

		if err != nil {
			return false, err
		}

		r := bytes.NewReader(b.Bytes())
		fh := ioutil.NopCloser(r)

		cache_fh, cache_err := c.Set(key, fh)

		if cache_err != nil {
			return false, cache_err
		}

		cache_err = cache_fh.Close()

		if cache_err != nil {
			return false, cache_err
		}

		return false, nil
	}

	if !cache.IsCacheMiss(err) {
		return false, err
	}

	return false, nil
}
