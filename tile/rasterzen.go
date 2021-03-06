package tile

import (
	"encoding/json"
	"github.com/go-spatial/geom/slippy"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"
	"strings"
)

type RasterzenOptions struct {
	Refresh bool `json:"refresh"`
}

func DefaultRasterzenOptions() (*RasterzenOptions, error) {

	opts := RasterzenOptions{
		Refresh: false,
	}

	return &opts, nil
}

func RasterzenOptionsFromString(body string) (*RasterzenOptions, error) {
	r := strings.NewReader(body)
	return RasterzenOptionsFromReader(r)
}

func RasterzenOptionsFromFile(path string) (*RasterzenOptions, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	return RasterzenOptionsFromReader(fh)
}

func RasterzenOptionsFromReader(fh io.Reader) (*RasterzenOptions, error) {

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	return RasterzenOptionsFromBytes(body)
}

func RasterzenOptionsFromBytes(body []byte) (*RasterzenOptions, error) {

	var rz_opts *RasterzenOptions

	err := json.Unmarshal(body, &rz_opts)

	if err != nil {
		return nil, err
	}

	return rz_opts, nil
}

func RenderRasterzenTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options, rz_opts *RasterzenOptions) (io.ReadCloser, error) {

	z := int(t.Z)
	x := int(t.X)
	y := int(t.Y)

	// key := fmt.Sprintf("%d/%d/%d.json", z, x, y)
	// nextzen_key := filepath.Join("nextzen", key)
	// rasterzen_key := filepath.Join("rasterzen", key)

	nextzen_key := CacheKeyForTile(t, "nextzen", "json")
	rasterzen_key := CacheKeyForTile(t, "rasterzen", "json")

	var nextzen_data io.ReadCloser   // stuff sent back from nextzen.org
	var rasterzen_data io.ReadCloser // stuff sent back from nextzen.org

	var err error

	if !rz_opts.Refresh {
		rasterzen_data, err = c.Get(rasterzen_key)

		if err == nil {
			return rasterzen_data, nil
		}
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

	// see notes in nextzen/tile.go about moving all the overzoom-ing
	// code in here which would allow us to pre-cache the Z16 tile...
	// (20190606/thisisaaronland)

	if nextzen.IsOverZoom(z) {
		rasterzen_data, err = c.Set(rasterzen_key, nextzen_data)
	} else {

		cr, err := nextzen.CropTile(z, x, y, nextzen_data)

		if err != nil {
			return nil, err
		}

		defer cr.Close()

		rasterzen_data, err = c.Set(rasterzen_key, cr)

		if err != nil {
			return nil, err
		}
	}

	return rasterzen_data, nil
}

func RasterzenToFeatureCollection(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

	features := make([]interface{}, 0)

	for _, l := range nextzen.Layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		gjson_features := fc.Get("features")

		for _, f := range gjson_features.Array() {
			features = append(features, f.Value())
		}
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	b, err := json.Marshal(fc)

	if err != nil {
		return err
	}

	_, err = out.Write(b)
	return err
}
