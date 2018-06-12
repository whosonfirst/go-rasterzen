package http

import (
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	_ "log"
	gohttp "net/http"
)

func GeoJSONHandler(c cache.Cache) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "text/json",
		"Access-Control-Allow-Origin": "*",
	}

	h := CacheHandler{
		Cache:   c,
		Func:    tile.ToFeatureCollection,
		Headers: headers,
	}

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		key := req.URL.Path
		err := h.HandleRequest(rsp, req, key)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	return gohttp.HandlerFunc(fn), nil
}
