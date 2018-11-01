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

	// it might seem weird and inefficient to be starting from SeedGeoJSON
	// rather than SeedSVG() and it probably is but the way the tile/*.go code
	// is written now coupled with the fact that oksvg doesn't accept io.Reader
	// thingies (only files) coupled with the fact that maybe you don't _want_
	// to have a non-ephemeral cache of the SVG renders (dunno, that's your
	// business) means that the code to work straight off an SVG cache looks
	// like the code that follows... so for now, we don't (20181101/thisisaaronland)

	/*

		svg_fh, err := SeedSVG(t, c, nz_opts)

		if err != nil {
			return nil, err
		}

		defer svg_fh.Close()

		tmpfile, err := ioutil.TempFile("", "svg")

		defer func() {

			_, err := os.Stat(tmpfile.Name())

			if !os.IsNotExist(err) {
				os.Remove(tmpfile.Name())
			}
		}()

		_, err = io.Copy(tmpfile, svg_fh)

		if err != nil {
			return nil, err
		}

		tmpfile.Close()

		// this doesn't even exist and would be in tile/geojson/go
		// which is... weird

		im, err := tile.ToImageFromSVG(tmpfile.Name())

		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		png_wr := bufio.NewWriter(&buf)

		err = png.Encode(png_wr, im)

		if err != nil {
			return nuil, err
		}

		png_wr.Flush()

		r := bytes.NewReader(buf.Bytes())
		png_fh := ioutil.NopCloser(r)

		return c.Set(png_key, png_fh)
	*/
}
