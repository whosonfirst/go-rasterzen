package tile

import (
	"bufio"
	"bytes"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
)

func RenderGeoJSONTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	geojson_key := CacheKeyForTile(t, "geojson", "geojson")

	var geojson_data io.ReadCloser
	var err error

	geojson_data, err = c.Get(geojson_key)

	if err == nil {
		return geojson_data, nil
	}

	rasterzen_fh, err := RenderRasterzenTile(t, c, nz_opts)

	if err != nil {
		return nil, err
	}

	defer rasterzen_fh.Close()

	var buf bytes.Buffer
	geojson_wr := bufio.NewWriter(&buf)

	err = RasterzenToGeoJSON(rasterzen_fh, geojson_wr)

	if err != nil {
		return nil, err
	}

	geojson_wr.Flush()

	r := bytes.NewReader(buf.Bytes())
	geojson_fh := ioutil.NopCloser(r)

	return c.Set(geojson_key, geojson_fh)
}
