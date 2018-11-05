package http

import (
	_ "github.com/whosonfirst/go-rasterzen/tile"
	"log"
	gohttp "net/http"
)

func RasterzenHandler(h *DispatchHandler) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "text/json",
		"Access-Control-Allow-Origin": "*",
	}

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
