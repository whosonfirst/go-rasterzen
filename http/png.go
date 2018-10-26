package http

import (
	"github.com/whosonfirst/go-rasterzen/tile"
	"log"
	gohttp "net/http"
)

func PNGHandler(h *CacheHandler) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "image/png",
		"Access-Control-Allow-Origin": "*",
	}

	h.Func = tile.ToPNG
	h.Headers = headers

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
