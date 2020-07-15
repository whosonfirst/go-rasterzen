package tile

import (
	"bytes"
	"encoding/json"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/paulmach/orb/maptile"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "log"
	"math"
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

/*

	the following methods were copied over from
	https://github.com/go-spatial/geom/blob/4a3ad4ba2b912d8dc0e39eedb0eb1b6b5e228d7d/slippy/tile.go

	because they were removed at some point and replaced
	with this:

	grid, err := slippy.NewGrid(4326)

	if err != nil {
		return err
	}

	ext, ok := slippy.Extent(grid, slippy_tile)

	if !ok {
		return errors.New("Unable to determine tile extent")
	}

	I can not figure out how to use these new methods to simply return
	a 4326 bounding box for a tile or more specifically a correct
	bounding box... (20200715/thisisaaronland)

	2020/07/15 13:52:24 TILE &{15 5242 12685}
	2020/07/15 13:52:24 GRID &{4326 2 180 90}
	2020/07/15 13:52:24 EXTENT &[-151.204833984375 20.313720703125 -151.1993408203125 20.3192138671875]

*/

func Tile2Lon(zoom, x uint) float64 {
	return float64(x)/math.Exp2(float64(zoom))*360.0 - 180.0
}

func Tile2Lat(zoom, y uint) float64 {
	var n float64 = math.Pi
	if y != 0 {
		n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(zoom))
	}

	return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}

func Extent4326(t *slippy.Tile) *geom.Extent {

	return geom.NewExtent(
		[2]float64{Tile2Lon(t.Z, t.X), Tile2Lat(t.Z, t.Y+1)},
		[2]float64{Tile2Lon(t.Z, t.X+1), Tile2Lat(t.Z, t.Y)},
	)
}
