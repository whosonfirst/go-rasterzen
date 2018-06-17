package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	_ "log"
)

type Cache interface {
	Get(string) (io.ReadCloser, error)
	Set(string, io.ReadCloser) (io.ReadCloser, error)
	Unset(string) error
	Hits() int64
	Misses() int64
	Evictions() int64
	Size() int64
}

type CacheMiss struct {
}

func (m CacheMiss) Error() string {
	return "MISS"
}

func IsCacheMiss(e error) bool {

	switch e.(type) {
	case *CacheMiss:
		return true
	case CacheMiss:
		return true
	default:
		// pass
	}

	return false
}

type BytesReadCloser struct {
	io.Reader
}

func (BytesReadCloser) Close() error { return nil }

func NewBytesReadCloserFromString(s string) BytesReadCloser {
	return NewBytesReadCloser([]byte(s))
}

func NewBytesReadCloser(body []byte) BytesReadCloser {
	r := bytes.NewReader(body)
	return BytesReadCloser{r}
}

func SetString(c Cache, k string, v string) (string, error) {

	r := NewBytesReadCloserFromString(v)
	fh, err := c.Set(k, r)

	if err != nil {
		return "", err
	}

	defer fh.Close()

	return toString(fh)
}

func GetString(c Cache, k string) (string, error) {

	fh, err := c.Get(k)

	if err != nil {
		return "", err
	}

	defer fh.Close()

	return toString(fh)
}

func toString(fh io.Reader) (string, error) {

	b, err := ioutil.ReadAll(fh)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
