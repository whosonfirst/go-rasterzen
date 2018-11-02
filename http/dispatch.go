package http

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
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

type DispatchFunc func(io.Reader, io.Writer) error

type DispatchHandler struct {
	Cache          cache.Cache
	Func           DispatchFunc
	Headers        map[string]string
	NextzenOptions *nextzen.Options
}

func NewDispatchHandler(c cache.Cache) (*DispatchHandler, error) {

	default_opts := new(nextzen.Options)

	default_headers := make(map[string]string)

	default_func := func(r io.Reader, wr io.Writer) error {
		_, err := io.Copy(wr, r)
		return err
	}

	h := DispatchHandler{
		Cache:          c,
		NextzenOptions: default_opts,
		Func:           default_func,
		Headers:        default_headers,
	}

	return &h, nil
}

func (h *DispatchHandler) HandleRequest(rsp gohttp.ResponseWriter, req *gohttp.Request, key string) error {

	data, err := h.Cache.Get(key)

	// log.Println("GET", key, err, cache.IsCacheMiss(err), cache.IsCacheMissMulti(err))

	if err == nil || cache.IsCacheMissMulti(err) {

		defer data.Close()

		// log.Printf("REQ %s RETURN FROM CACHE\n", key)

		for k, v := range h.Headers {
			rsp.Header().Set(k, v)
		}

		if !cache.IsCacheMissMulti(err) {
			_, err = io.Copy(rsp, data)
			return err
		}

		var b bytes.Buffer
		buf := bufio.NewWriter(&b)

		wr := io.MultiWriter(rsp, buf)

		_, err = io.Copy(wr, data)

		buf.Flush()

		if err == nil {

			_, cache_err := h.Cache.Set(key, cache.NewReadCloser(b.Bytes()))

			if cache_err != nil {
				log.Printf("%s %v\n", key, cache_err)
			}
		}

		// log.Printf("REQ %s UPDATE CACHE %v\n", key, err)
		return err
	}

	if err != nil && !cache.IsCacheMiss(err) {
		return err
	}

	t, err := h.GetSlippyTileForRequest(req)

	if err != nil {
		return err
	}

	nz_opts := h.NextzenOptions

	if nz_opts.ApiKey == "" {

		url := req.URL
		query := url.Query()

		api_key := query.Get("api_key")

		if api_key == "" {
			return errors.New("Missing API key")
		}

		nz_opts.ApiKey = api_key
	}

	fh, err := seed.SeedRasterzen(*t, h.Cache, nz_opts)

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

	// this is the thing that transforms the rasterzen
	// tile in to geojson, svg, png, etc.

	err = h.Func(fh, wr)

	buf.Flush()

	if err != nil {
		return err
	}

	_, cache_err := h.Cache.Set(key, cache.NewReadCloser(b.Bytes()))

	if cache_err != nil {
		log.Printf("%s %v\n", key, cache_err)
	}

	return nil
}

func (h DispatchHandler) GetSlippyTileForRequest(req *gohttp.Request) (*slippy.Tile, error) {

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

	t := slippy.Tile{
		Z: uint(z),
		X: uint(x),
		Y: uint(y),
	}

	return &t, nil
}
