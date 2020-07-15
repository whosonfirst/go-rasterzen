package http

import (
	"errors"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"log"
	gohttp "net/http"
)

func NewSVGHandler(c cache.Cache, nz_opts *nextzen.Options, svg_opts *tile.RasterzenSVGOptions) (gohttp.HandlerFunc, error) {

	d, err := NewDispatchHandler(c)

	if err != nil {
		return nil, err
	}

	d.NextzenOptions = nz_opts
	d.SVGOptions = svg_opts

	return SVGHandler(d)
}

func SVGHandler(h *DispatchHandler) (gohttp.HandlerFunc, error) {

	headers := map[string]string{
		"Content-Type":                "image/svg+xml",
		"Access-Control-Allow-Origin": "*",
	}

	h.Func = func(slippy_tile *slippy.Tile, in io.Reader, out io.Writer) error {
		
		svg_opts := h.SVGOptions.Clone()

		// svg_opts.TileExtent = slippy_tile.Extent4326()

		grid, err := slippy.NewGrid(4326)

		if err != nil {
			return err
		}

		ext, ok := slippy.Extent(grid, slippy_tile)

		if !ok {
			return errors.New("Unable to determine tile extent")
		}

		log.Println("TILE", slippy_tile)
		log.Println("GRID", grid)
		log.Println("EXTENT", ext)
		
		svg_opts.TileExtent = ext

		return tile.RasterzenToSVGWithOptions(in, out, svg_opts)
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
