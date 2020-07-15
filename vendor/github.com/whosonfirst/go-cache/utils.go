package cache

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
)

func NewReadCloser(b []byte) io.ReadCloser {
	r := bytes.NewReader(b)
	return ioutil.NopCloser(r)
}

func NewReadCloserFromString(s string) io.ReadCloser {
	return NewReadCloser([]byte(s))
}

func SetString(c Cache, k string, v string) (string, error) {

	ctx := context.Background()

	r := NewReadCloserFromString(v)
	fh, err := c.Set(ctx, k, r)

	if err != nil {
		return "", err
	}

	defer fh.Close()

	return toString(fh)
}

func GetString(c Cache, k string) (string, error) {

	ctx := context.Background()

	fh, err := c.Get(ctx, k)

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
