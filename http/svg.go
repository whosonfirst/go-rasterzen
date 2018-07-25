package http

import (
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"log"
	gohttp "net/http"
)

func SVGHandler(c cache.Cache) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "image/svg+xml",
		"Access-Control-Allow-Origin": "*",
	}

	h := CacheHandler{
		Cache:   c,
		Func:    tile.ToSVG,
		Headers: headers,
	}

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		key := req.URL.Path
		err := h.HandleRequest(rsp, req, key)

		if err != nil {
			log.Printf("%s %v\n", key, err)
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	return gohttp.HandlerFunc(fn), nil
}
