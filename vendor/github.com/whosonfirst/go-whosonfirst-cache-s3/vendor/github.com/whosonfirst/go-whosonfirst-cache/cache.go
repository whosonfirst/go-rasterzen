package cache

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	_ "log"
)

type Cache interface {
	Name() string
	Get(string) (io.ReadCloser, error)
	Set(string, io.ReadCloser) (io.ReadCloser, error)
	Unset(string) error
	Hits() int64
	Misses() int64
	Evictions() int64
	Size() int64
}

type CacheMiss struct {
	error string
}

func (m CacheMiss) Error() string {

	return fmt.Sprintf("CACHE MISS %s", m.error)
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

func NewReadCloser(b []byte) io.ReadCloser {
	r := bytes.NewReader(b)
	return ioutil.NopCloser(r)
}

func NewReadCloserFromString(s string) io.ReadCloser {
	return NewReadCloser([]byte(s))
}

func SetString(c Cache, k string, v string) (string, error) {

	r := NewReadCloserFromString(v)
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
