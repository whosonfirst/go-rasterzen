package nextzen

import (
	"bytes"
	"fmt"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	_ "log"
	"net/http"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

// PLEASE CACHE ME...

func FetchTile(z int, x int, y int, api_key string) (io.ReadCloser, error) {

	layer := "all"

	url := fmt.Sprintf("https://tile.nextzen.org/tilezen/vector/v1/256/%s/%d/%d/%d.json?api_key=%s", layer, z, x, y, api_key)

	rsp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	fh := rsp.Body
	defer fh.Close()

	return CropTile(z, x, y, fh)
}

func CropTile(z int, x int, y int, fh io.ReadCloser) (io.ReadCloser, error) {

	zm := maptile.Zoom(uint32(z))
	tl := maptile.Tile{uint32(x), uint32(y), zm}

	bounds := tl.Bound()

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	// PLEASE RECONCILE ME WITH tile/geojson.go

	layers := []string{
		"boundaries",
		"buildings",
		"earth",
		"landuse",
		"places",
		"pois",
		"roads",
		"transit",
		"water",
	}

	for _, l := range layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		features := fc.Get("features")

		// SOMETHING SOMETHING SOMETHING DO THESE ALL IN PARALLEL...

		for i, f := range features.Array() {

			str_f := f.String()

			feature, err := geojson.UnmarshalFeature([]byte(str_f))

			if err != nil {
				return nil, err
			}

			geom := feature.Geometry

			orb_geom := clip.Geometry(bounds, geom)
			new_geom := geojson.NewGeometry(orb_geom)

			path := fmt.Sprintf("%s.features.%d.geometry", l, i)
			body, err = sjson.SetBytes(body, path, new_geom)

			if err != nil {
				return nil, err
			}
		}
	}

	r := bytes.NewReader(body)
	return nopCloser{r}, nil
}
