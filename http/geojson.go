package http

import (
	"github.com/whosonfirst/go-rasterzen/tile"
	gohttp "net/http"
)

func GeoJSONHandler() (gohttp.HandlerFunc, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		fh, err := GetTileForRequest(req)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		defer fh.Close()

		rsp.Header().Set("Content-Type", "text/json")
		rsp.Header().Set("Access-Control-Allow-Origin", "*")

		err = tile.ToFeatureCollection(fh, rsp)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	return gohttp.HandlerFunc(fn), nil
}
