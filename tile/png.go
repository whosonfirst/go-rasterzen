package tile

import (
	"bufio"
	"bytes"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"image/png"
	"io"
	"io/ioutil"
)

func RenderPNGTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options, svg_opts *RasterzenSVGOptions) (io.ReadCloser, error) {

	png_key := CacheKeyForTile(t, "png", "png")

	var png_data io.ReadCloser
	var err error

	png_data, err = c.Get(png_key)

	if err == nil {
		return png_data, nil
	}

	rasterzen_fh, err := RenderRasterzenTile(t, c, nz_opts)

	if err != nil {
		return nil, err
	}

	defer rasterzen_fh.Close()

	var buf bytes.Buffer
	png_wr := bufio.NewWriter(&buf)

	err = RasterzenToPNG(rasterzen_fh, png_wr, svg_opts)

	if err != nil {
		return nil, err
	}

	png_wr.Flush()

	r := bytes.NewReader(buf.Bytes())
	png_fh := ioutil.NopCloser(r)

	return c.Set(png_key, png_fh)
}

func RasterzenToPNG(in io.Reader, out io.Writer, svg_opts *RasterzenSVGOptions) error {

	img, err := RasterzenToImageWithOptions(in, svg_opts)

	if err != nil {
		return err
	}

	return png.Encode(out, img)
}
