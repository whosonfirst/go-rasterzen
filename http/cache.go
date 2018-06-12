package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"log"
	gohttp "net/http"
	"path/filepath"
	"regexp"
	"strconv"
)

var re_path *regexp.Regexp

func init() {
	re_path = regexp.MustCompile(`/(.*)/(\d+)/(\d+)/(\d+).(\w+)$`)
}

type CacheHandlerFunc func(io.Reader, io.Writer) error

type CacheHandler struct {
	Cache   cache.Cache
	Func    CacheHandlerFunc
	Headers map[string]string
}

func (h CacheHandler) HandleRequest(rsp gohttp.ResponseWriter, req *gohttp.Request, key string) error {

	data, err := h.Cache.Get(key)

	if err == nil {

		defer data.Close()

		for k, v := range h.Headers {
			rsp.Header().Set(k, v)
		}

		_, err = io.Copy(rsp, data)
		return err
	}

	if err != nil && !cache.IsCacheMiss(err) {
		log.Println("CACHE ERROR", key, err)
	}

	fh, err := h.GetTileForRequest(req)

	if err != nil {
		return err
	}

	defer fh.Close()

	for k, v := range h.Headers {
		rsp.Header().Set(k, v)
	}

	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	wr := io.MultiWriter(rsp, buf)

	err = h.Func(fh, wr)

	buf.Flush()

	if err != nil {
		return err
	}

	go h.Cache.Set(key, cache.NewBytesReadCloser(b.Bytes()))
	return nil
}

func (h CacheHandler) GetTileForRequest(req *gohttp.Request) (io.ReadCloser, error) {

	path := req.URL.Path

	if !re_path.MatchString(path) {
		return nil, errors.New("Invalid path")
	}

	m := re_path.FindStringSubmatch(path)

	z, err := strconv.Atoi(m[2])

	if err != nil {
		return nil, err
	}

	x, err := strconv.Atoi(m[3])

	if err != nil {
		return nil, err
	}

	y, err := strconv.Atoi(m[4])

	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%d/%d/%d.json", z, x, y)

	nextzen_key := filepath.Join("nextzen", key)
	rasterzen_key := filepath.Join("rasterzen", key)

	var nextzen_data io.ReadCloser   // stuff sent back from nextzen.org
	var rasterzen_data io.ReadCloser // nextzen.org data cropped and manipulated

	rasterzen_data, err = h.Cache.Get(rasterzen_key)

	if err == nil {
		return rasterzen_data, nil
	}

	nextzen_data, err = h.Cache.Get(nextzen_key)

	if err != nil {

		url := req.URL
		query := url.Query()

		api_key := query.Get("api_key")

		if api_key == "" {
			return nil, errors.New("Missing API key")
		}

		t, err := nextzen.FetchTile(z, x, y, api_key)

		if err != nil {
			return nil, err
		}

		defer t.Close()

		nextzen_data, err = h.Cache.Set(nextzen_key, t)

		if err != nil {
			return nil, err
		}
	}

	cr, err := nextzen.CropTile(z, x, y, nextzen_data)

	if err != nil {
		return nil, err
	}

	defer cr.Close()

	return h.Cache.Set(rasterzen_key, cr)
}
