package http

import (
	"github.com/whosonfirst/go-rasterzen/mvt"
	gohttp "net/http"
)

func PNGHandler() (gohttp.HandlerFunc, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		fh, err := GetTileForRequest(req)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		defer fh.Close()

		rsp.Header().Set("Content-Type", "image/png")

		err = mvt.ToPNG(fh, rsp)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	return gohttp.HandlerFunc(fn), nil
}
