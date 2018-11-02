package seed

import (
	"errors"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-log"
	"sync"
	"sync/atomic"
	"time"
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

	if t.Z > 16 {
		return errors.New("Tiles > zoom 16 can not be rendered at this time.")
	}

	k := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
	ts.tile_map.LoadOrStore(k, t)
	return nil
}

func (ts *TileSet) Range(f func(key, value interface{}) bool) {
	ts.tile_map.Range(f)
}

func (ts *TileSet) Count() int32 {

	remaining := int32(0)

	tile_func := func(key, value interface{}) bool {
		atomic.AddInt32(&remaining, 1)
		return true
	}

	ts.Range(tile_func)

	return remaining
}

type TileSeeder struct {
	Cache          cache.Cache
	Seeders        int
	NextzenOptions *nextzen.Options
	SeedSVG        bool
	SeedPNG        bool
	Timings        bool
	Logger         *log.WOFLogger
}

func NewTileSeeder(c cache.Cache, nz_opts *nextzen.Options) (*TileSeeder, error) {

	logger := log.SimpleWOFLogger()

	s := TileSeeder{
		Cache:          c,
		NextzenOptions: nz_opts,
		SeedSVG:        true,
		SeedPNG:        false,
		Seeders:        100,
		Timings:        false,
		Logger:         logger,
	}

	return &s, nil
}

func (s *TileSeeder) SeedTileSet(ts *TileSet) (bool, []error) {

	if s.Timings {

		t1 := time.Now()

		defer func() {
			s.Logger.Status("Time to seed all tiles %v\n", time.Since(t1))
		}()
	}

	throttle := make(chan bool, s.Seeders)

	for i := 0; i < s.Seeders; i++ {
		throttle <- true
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)

	remaining := ts.Count()

	tile_func := func(key, value interface{}) bool {

		t := value.(slippy.Tile)

		go func(t slippy.Tile, throttle chan bool, done_ch chan bool, err_ch chan error) {

			<-throttle

			if s.Timings {

				t1 := time.Now()

				defer func() {
					s.Logger.Status("Time to seed tile (%v) %v\n", t, time.Since(t1))
				}()
			}

			defer func() {
				done_ch <- true
				throttle <- true
			}()

			err := s.SeedTile(t)

			if err != nil {
				msg := fmt.Sprintf("Unabled to seed %v because %s", t, err)
				err_ch <- errors.New(msg)
				return
			}

		}(t, throttle, done_ch, err_ch)

		return true
	}

	ts.Range(tile_func)

	errors := make([]error, 0)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			errors = append(errors, e)
		default:
			//
		}
	}

	ok := len(errors) == 0
	return ok, errors
}

// this is basically the http/cache.go GetTileForRequest() function so once we
// have it working here then we should reconcile the two pieces of code...
// (20181101/thisisaaronland)

// something something something what to do about SVG and PNG tiles?
// (20181101/thisisaaronland)

func (s *TileSeeder) SeedTile(t slippy.Tile) error {

	if !s.SeedSVG && !s.SeedPNG {

		_, err := SeedRasterzen(t, s.Cache, s.NextzenOptions)

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
