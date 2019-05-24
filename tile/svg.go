package tile

import (
	"bufio"
	"bytes"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "path/filepath"
)

func RenderSVGTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options, svg_opts *RasterzenSVGOptions) (io.ReadCloser, error) {

	svg_key := CacheKeyForTile(t, "svg", "svg")

	var svg_data io.ReadCloser
	var err error

	svg_data, err = c.Get(svg_key)

	if err == nil {
		return svg_data, nil
	}

	geojson_fh, err := RenderRasterzenTile(t, c, nz_opts)

	if err != nil {
		return nil, err
	}

	defer geojson_fh.Close()

	var buf bytes.Buffer
	svg_wr := bufio.NewWriter(&buf)

	err = RasterzenToSVGWithOptions(geojson_fh, svg_wr, svg_opts)

	if err != nil {
		return nil, err
	}

	svg_wr.Flush()

	r := bytes.NewReader(buf.Bytes())
	svg_fh := ioutil.NopCloser(r)

	return c.Set(svg_key, svg_fh)
}
