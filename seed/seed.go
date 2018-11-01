package seed

import (
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
)

type TileSower struct {
	Cache          cache.Cache
	NextzenOptions *nextzen.Options
	SeedGeoJSON    bool
	SeedSVG        bool
	SeedPNG        bool
}

func NewTileSower(c cache.Cache, nz_opts *nextzen.Options) (*TileSower, error) {

	s := TileSower{
		Cache:          c,
		NextzenOptions: nz_opts,
		SeedSVG:        true,
		SeedPNG:        false,
	}

	return &s, nil
}

// this is basically the http/cache.go GetTileForRequest() function so once we
// have it working here then we should reconcile the two pieces of code...
// (20181101/thisisaaronland)

// something something something what to do about SVG and PNG tiles?
// (20181101/thisisaaronland)

func (s *TileSower) SeedTile(t slippy.Tile) error {

	if !s.SeedSVG && !s.SeedPNG {

		_, err := SeedGeoJSON(t, s.Cache, s.NextzenOptions)

		if err != nil {
			return err
		}
	}

	if s.SeedSVG {

		_, err := SeedSVG(t, s.Cache, s.NextzenOptions)

		if err != nil {
			return err
		}

	}

	if s.SeedPNG {

		_, err := SeedPNG(t, s.Cache, s.NextzenOptions)

		if err != nil {
			return err
		}

	}

	return nil
}
