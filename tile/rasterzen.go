package tile

import (
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"path/filepath"
)

func SeedRasterzen(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	z := int(t.Z)
	x := int(t.X)
	y := int(t.Y)

	key := fmt.Sprintf("%d/%d/%d.json", z, x, y)

	nextzen_key := filepath.Join("nextzen", key)
	rasterzen_key := filepath.Join("rasterzen", key)

	var nextzen_data io.ReadCloser   // stuff sent back from nextzen.org
	var rasterzen_data io.ReadCloser // stuff sent back from nextzen.org

	var err error

	rasterzen_data, err = c.Get(rasterzen_key)

	if err == nil {
		return rasterzen_data, nil
	}

	nextzen_data, err = c.Get(nextzen_key)

	if err != nil {

		t, err := nextzen.FetchTile(z, x, y, nz_opts)

		if err != nil {
			return nil, err
		}

		defer t.Close()

		nextzen_data, err = c.Set(nextzen_key, t)

		if err != nil {
			return nil, err
		}
	}

	cr, err := nextzen.CropTile(z, x, y, nextzen_data)

	if err != nil {
		return nil, err
	}

	defer cr.Close()

	rasterzen_data, err = c.Set(rasterzen_key, cr)

	if err != nil {
		return nil, err
	}

	return rasterzen_data, nil
}
