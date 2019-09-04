package tile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/go-spatial/geom/slippy"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
)

func RenderGeoJSONTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options, rz_opts *RasterzenOptions) (io.ReadCloser, error) {

	geojson_key := CacheKeyForTile(t, "geojson", "geojson")

	var geojson_data io.ReadCloser
	var err error

	geojson_data, err = c.Get(geojson_key)

	if err == nil {
		return geojson_data, nil
	}

	rasterzen_fh, err := RenderRasterzenTile(t, c, nz_opts, rz_opts)

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

func RasterzenToGeoJSON(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

	type FeatureCollection struct {
		Type     string        `json:"type"`
		Features []interface{} `json:"features"`
	}

	features := make([]interface{}, 0)

	for _, l := range nextzen.Layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		fc_features := fc.Get("features")

		for _, f := range fc_features.Array() {
			features = append(features, f.Value())
		}
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	enc, err := json.Marshal(fc)

	if err != nil {
		return err
	}

	r := bytes.NewReader(enc)
	_, err = io.Copy(out, r)

	if err != nil {
		return err
	}

	return nil
}
