package http

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"log"
	gohttp "net/http"
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

		log.Println("CACHE HIT", key)
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

	log.Println("CACHE MISS", key)

	fh, err := GetTileForRequest(req)

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

	log.Println("CACHE SET", key)
	go h.Cache.Set(key, cache.NewBytesReadCloser(b.Bytes()))
	return nil
}

func GetTileForRequest(req *gohttp.Request) (io.ReadCloser, error) {

	url := req.URL
	path := url.Path

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

	query := url.Query()
	api_key := query.Get("api_key")

	if api_key == "" {
		return nil, errors.New("Missing API key")
	}

	return nextzen.FetchTile(z, x, y, api_key)
}
