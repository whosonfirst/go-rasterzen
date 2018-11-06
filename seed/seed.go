package seed

import (
	"errors"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/worker"
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
	worker        worker.Worker
	MaxWorkers    int
	SeedRasterzen bool
	SeedSVG       bool
	SeedPNG       bool
	Timings       bool
	Logger        *log.WOFLogger
}

func NewTileSeeder(w worker.Worker) (*TileSeeder, error) {

	logger := log.SimpleWOFLogger()

	s := TileSeeder{
		worker:        w,
		SeedRasterzen: true,
		SeedSVG:       true,
		SeedPNG:       false,
		MaxWorkers:    100,
		Timings:       false,
		Logger:        logger,
	}

	return &s, nil
}

func (s *TileSeeder) SeedTileSet(ts *TileSet) (bool, []error) {

	if s.Timings {

		t1 := time.Now()

		defer func() {
			s.Logger.Status("Time to seed all tiles %v", time.Since(t1))
		}()
	}

	throttle := make(chan bool, s.MaxWorkers)

	for i := 0; i < s.MaxWorkers; i++ {
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
					s.Logger.Status("Time to seed tile (%v) %v", t, time.Since(t1))
				}()
			}

			defer func() {
				done_ch <- true
				throttle <- true
			}()

			ok, errs := s.seedTiles(t)

			if !ok {

				for _, e := range errs {
					msg := fmt.Sprintf("Unabled to seed %v because %s", t, e)
					err_ch <- errors.New(msg)
				}
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

func (s *TileSeeder) seedTiles(t slippy.Tile) (bool, []error) {

	if s.SeedRasterzen {

		// please figure me out... (20181105/thisisaaronland)

		/*
			cache_key := tile.CacheKeyForRasterzenTile(t)

			cached, err := worker.IsCached(s.worker.cache, cache_key)

			if err != nil {
				return err
			}

			if !cached {
		*/

		err := s.worker.RenderRasterzenTile(t)

		if err != nil {
			return false, []error{err}
		}
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)

	remaining := 0

	if s.SeedPNG {

		remaining += 1

		go func() {
			err := s.worker.RenderPNGTile(t)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true
		}()
	}

	if s.SeedSVG {

		remaining += 1

		go func() {

			err := s.worker.RenderSVGTile(t)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true
		}()
	}

	errors := make([]error, 0)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			errors = append(errors, e)
		default:
			// pass
		}
	}

	ok := len(errors) == 0
	return ok, errors
}
