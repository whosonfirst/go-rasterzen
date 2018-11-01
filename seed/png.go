package seed

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "path/filepath"
)

func SeedPNG(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	z := int(t.Z)
	x := int(t.X)
	y := int(t.Y)

	png_key := fmt.Sprintf("png/%d/%d/%d.png", z, x, y)

	var png_data io.ReadCloser
	var err error

	png_data, err = c.Get(png_key)

	if err == nil {
		return png_data, nil
	}

	geojson_fh, err := SeedGeoJSON(t, c, nz_opts)

	if err != nil {
		return nil, err
	}

	defer geojson_fh.Close()

	var buf bytes.Buffer
	png_wr := bufio.NewWriter(&buf)

	err = tile.ToPNG(geojson_fh, png_wr)

	if err != nil {
		return nil, err
	}

	png_wr.Flush()

	r := bytes.NewReader(buf.Bytes())
	png_fh := ioutil.NopCloser(r)

	return c.Set(png_key, png_fh)
}
