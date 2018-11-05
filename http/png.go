package http

import (
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"log"
	gohttp "net/http"
)

func NewPNGHandler(c cache.Cache, nz_opts *nextzen.Options) (gohttp.HandlerFunc, error) {

	d, err := NewDispatchHandler(c)

	if err != nil {
		return nil, err
	}

	d.NextzenOptions = nz_opts
	return PNGHandler(d)
}

func PNGHandler(h *DispatchHandler) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "image/png",
		"Access-Control-Allow-Origin": "*",
	}

	h.Func = tile.RasterzenToPNG
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
