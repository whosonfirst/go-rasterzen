package http

import (
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	gohttp "net/http"
)

func PNGHandler(c cache.Cache) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "image/png",
		"Access-Control-Allow-Origin": "*",
	}

	h := CacheHandler{
		Cache:   c,
		Func:    tile.ToPNG,
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
