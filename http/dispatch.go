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

	if err != nil {
		return err
	}

	t, err := h.GetSlippyTileForRequest(req)

	if err != nil {
		return err
	}

	// something something something API key

	fh, err := seed.SeedGeoJSON(*t, h.Cache, h.NextzenOptions)

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

// deprecated

/*
func (h DispatchHandler) GetTileForRequest(req *gohttp.Request) (io.ReadCloser, error) {

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

	// this is the new new (I think...) but it doesn't work yet
	// return rasterzen.GetTileWithCache(h.Cache, z, x, y)

	key := fmt.Sprintf("%d/%d/%d.json", z, x, y)

	nextzen_key := filepath.Join("nextzen", key)
	rasterzen_key := filepath.Join("rasterzen", key)

	var nextzen_data io.ReadCloser   // stuff sent back from nextzen.org
	var rasterzen_data io.ReadCloser // nextzen.org data cropped and manipulated

	rasterzen_data, err = h.Cache.Get(rasterzen_key)

	// log.Printf("REQ RASTERZEN %s %v\n", rasterzen_key, err)

	if err == nil {
		return rasterzen_data, nil
	}

	nextzen_data, err = h.Cache.Get(nextzen_key)

	// log.Printf("REQ NEXTZEN %s %v\n", nextzen_key, err)

	if err != nil {

		opts := h.NextzenOptions

		url := req.URL
		query := url.Query()

		if opts.ApiKey == "" {

			api_key := query.Get("api_key")

			if api_key == "" {
				return nil, errors.New("Missing API key")
			}

			opts.ApiKey = api_key
		}

		// check for and set 'Origin' header?

		t, err := nextzen.FetchTile(z, x, y, opts)

		// log.Printf("REQ NEXTZEN %d/%d/%d %v\n", z, x, y, err)

		if err != nil {
			return nil, err
		}

		defer t.Close()

		nextzen_data, err = h.Cache.Set(nextzen_key, t)

		// log.Printf("REQ NEXTZEN SET %s %v\n", nextzen_key, err)

		if err != nil {
			return nil, err
		}
	}

	cr, err := nextzen.CropTile(z, x, y, nextzen_data)

	// log.Printf("REQ NEXTZEN CROP %d/%d/%d %v\n", z, x, y, err)

	if err != nil {
		return nil, err
	}

	defer cr.Close()

	fh, err := h.Cache.Set(rasterzen_key, cr)

	// log.Printf("REQ RASTERZEN SET %s %v\n", rasterzen_key, err)

	return fh, err
}

*/
