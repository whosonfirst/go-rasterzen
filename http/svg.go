package http

import (
	"github.com/whosonfirst/go-rasterzen/mvt"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	gohttp "net/http"
	"regexp"
	"strconv"
)

func SVGHandler(api_key string) (gohttp.HandlerFunc, error) {

	re, err := regexp.Compile(`/(.*)/(\d+)/(\d+)/(\d+).(\w+)$`)

	if err != nil {
		return nil, err
	}

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		url := req.URL
		path := url.Path

		if !re.MatchString(path) {
			gohttp.Error(rsp, "404 Not found", gohttp.StatusNotFound)
			return
		}

		m := re.FindStringSubmatch(path)

		z, err := strconv.Atoi(m[2])

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		x, err := strconv.Atoi(m[3])

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		y, err := strconv.Atoi(m[4])

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		fh, err := nextzen.FetchTile(z, x, y, api_key)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		defer fh.Close()

		rsp.Header().Set("Content-Type", "image/svg+xml")
		rsp.Header().Set("Access-Control-Allow-Origin", "*")

		err = mvt.ToSVG(fh, rsp)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	return gohttp.HandlerFunc(fn), nil
}
