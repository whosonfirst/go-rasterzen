package tile

import (
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
)

func RenderGeoJSONTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	geojson_key := CacheKeyForTile(t, "geojson", "geojson")

	var geojson_data io.ReadCloser
	var err error

	geojson_data, err = c.Get(geojson_key)

	if err == nil {
		return geojson_data, nil
	}

	geojson_fh, err := RenderRasterzenTile(t, c, nz_opts)

	if err != nil {
		return nil, err
	}

	defer geojson_fh.Close()

	return c.Set(geojson_key, geojson_fh)
}
