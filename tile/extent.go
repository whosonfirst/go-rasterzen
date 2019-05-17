package tile

import (
	"bytes"
	"encoding/json"
	"github.com/go-spatial/geom/slippy"
	"github.com/paulmach/orb/maptile"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "log"
)

type Feature struct {
	Type       string            `json:"type"`
	Geometry   Geometry          `json:"geometry"`
	Properties map[string]string `json:"properties"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

type Coordinate []float64

type Coordinates []Coordinate

type Polygon []Coordinates

func RenderExtentTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	extent_key := CacheKeyForTile(t, "extent", "geojson")

	var extent_data io.ReadCloser
	var err error

	extent_data, err = c.Get(extent_key)

	if err == nil {
		return extent_data, nil
	}

	z := uint32(t.Z)
	x := uint32(t.X)
	y := uint32(t.Y)

	zm := maptile.Zoom(z)
	tl := maptile.New(x, y, zm)

	bounds := tl.Bound()

	minx := bounds.Min.Lon()
	miny := bounds.Min.Lat()
	maxx := bounds.Max.Lon()
	maxy := bounds.Max.Lat()

	coords := []Coordinate{
		Coordinate{minx, miny},
		Coordinate{minx, maxy},
		Coordinate{maxx, maxy},
		Coordinate{maxx, miny},
		Coordinate{minx, miny},
	}

	poly := Polygon{coords}

	geom := Geometry{
		Type:        "Polygon",
		Coordinates: poly,
	}

	props := make(map[string]string)

	feature := Feature{
		Type:       "Feature",
		Geometry:   geom,
		Properties: props,
	}

	body, err := json.Marshal(feature)

	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(body)
	extent_fh := ioutil.NopCloser(r)

	return c.Set(extent_key, extent_fh)
}
