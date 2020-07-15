package tile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type RasterzenPNGOptions struct {
	Refresh bool `json:"refresh"`
}

func DefaultRasterzenPNGOptions() (*RasterzenPNGOptions, error) {

	opts := RasterzenPNGOptions{
		Refresh: false,
	}

	return &opts, nil
}

func RasterzenPNGOptionsFromString(body string) (*RasterzenPNGOptions, error) {
	r := strings.NewReader(body)
	return RasterzenPNGOptionsFromReader(r)
}

func RasterzenPNGOptionsFromFile(path string) (*RasterzenPNGOptions, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	return RasterzenPNGOptionsFromReader(fh)
}

func RasterzenPNGOptionsFromReader(fh io.Reader) (*RasterzenPNGOptions, error) {

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	return RasterzenPNGOptionsFromBytes(body)
}

func RasterzenPNGOptionsFromBytes(body []byte) (*RasterzenPNGOptions, error) {

	var png_opts *RasterzenPNGOptions

	err := json.Unmarshal(body, &png_opts)

	if err != nil {
		return nil, err
	}

	return png_opts, nil
}

func RenderPNGTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options, rz_opts *RasterzenOptions, svg_opts *RasterzenSVGOptions, png_opts *RasterzenPNGOptions) (io.ReadCloser, error) {

	png_key := CacheKeyForTile(t, "png", "png")

	var png_data io.ReadCloser
	var err error

	if !png_opts.Refresh {

		png_data, err = c.Get(png_key)

		if err == nil {
			return png_data, nil
		}
	}

	rasterzen_fh, err := RenderRasterzenTile(t, c, nz_opts, rz_opts)

	if err != nil {
		return nil, err
	}

	defer rasterzen_fh.Close()

	var buf bytes.Buffer
	png_wr := bufio.NewWriter(&buf)

	svg_opts.TileExtent = Extent4326(&t)

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
