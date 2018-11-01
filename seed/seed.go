package seed

import (
	"errors"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"log"
	"sync"
	"sync/atomic"
)

type TileSet struct {
	tile_map *sync.Map
}

func NewTileSet() (*TileSet, error) {

	tm := new(sync.Map)

	ts := TileSet{
		tile_map: tm,
	}

	return &ts, nil
}

func (ts *TileSet) AddTile(t slippy.Tile) error {

	k := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
	ts.tile_map.LoadOrStore(k, t)
	return nil
}

func (ts *TileSet) Range(f func(key, value interface{}) bool) {
	ts.tile_map.Range(f)
}

type TileSeeder struct {
	Cache          cache.Cache
	NextzenOptions *nextzen.Options
	SeedGeoJSON    bool
	SeedSVG        bool
	SeedPNG        bool
}

func NewTileSeeder(c cache.Cache, nz_opts *nextzen.Options) (*TileSeeder, error) {

	s := TileSeeder{
		Cache:          c,
		NextzenOptions: nz_opts,
		SeedSVG:        true,
		SeedPNG:        false,
	}

	return &s, nil
}

// TO DO: figure out how to handle errors...

func (s *TileSeeder) SeedTileSet(ts *TileSet) error {

	done_ch := make(chan bool)
	err_ch := make(chan error)

	remaining := int32(0)

	tile_func := func(key, value interface{}) bool {

		atomic.AddInt32(&remaining, 1)

		t := value.(slippy.Tile)

		go func(t slippy.Tile) {

			defer func() {
				done_ch <- true
			}()

			err := s.SeedTile(t)

			if err != nil {
				msg := fmt.Sprintf("Unabled to seed %v because %s", t, err)
				err_ch <- errors.New(msg)
				return
			}

			log.Println("OK", t)
		}(t)

		return true
	}

	ts.Range(tile_func)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			log.Println(e)
		default:
			//
		}
	}

	return nil
}

// this is basically the http/cache.go GetTileForRequest() function so once we
// have it working here then we should reconcile the two pieces of code...
// (20181101/thisisaaronland)

// something something something what to do about SVG and PNG tiles?
// (20181101/thisisaaronland)

func (s *TileSeeder) SeedTile(t slippy.Tile) error {

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
