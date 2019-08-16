package seed

import (
	"context"
	"errors"
	"fmt"
	"github.com/aaronland/go-string/dsn"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/seed/catalog"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-log"
	golog "log"
	"strings"
	"sync/atomic"
	"time"
)

type TileSet struct {
	tile_catalog catalog.SeedCatalog
}

func NewTileSetFromDSN(str_dsn string) (*TileSet, error) {

	dsn_map, err := dsn.StringToDSNWithKeys(str_dsn, "catalog")

	if err != nil {
		return nil, err
	}

	var seed_catalog catalog.SeedCatalog

	switch strings.ToUpper(dsn_map["catalog"]) {
	case "SQLITE":

		sqlite_dsn, ok := dsn_map["dsn"]

		if ok {
			seed_catalog, err = catalog.NewSQLiteSeedCatalog(sqlite_dsn)
		} else {
			err = errors.New("Missing 'dsn' property")
		}

	case "SYNC":
		seed_catalog, err = catalog.NewSyncMapSeedCatalog()
	default:
		err = errors.New("Invalid catalog")
	}

	if err != nil {
		return nil, err
	}

	return NewTileSet(seed_catalog)
}

func NewTileSet(seed_catalog catalog.SeedCatalog) (*TileSet, error) {

	ts := TileSet{
		tile_catalog: seed_catalog,
	}

	return &ts, nil
}

func (ts *TileSet) AddTile(t slippy.Tile) error {
	k := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
	ts.tile_catalog.LoadOrStore(k, t)
	return nil
}

func (ts *TileSet) Range(f func(key, value interface{}) bool) {
	ts.tile_catalog.Range(f)
}

func (ts *TileSet) Count() int32 {
	return ts.tile_catalog.Count()
}

type TileSeeder struct {
	worker        worker.Worker
	cache         cache.Cache
	MaxWorkers    int
	SeedRasterzen bool
	SeedGeoJSON   bool
	SeedExtent    bool
	SeedSVG       bool
	SeedPNG       bool
	SVGOptions    *tile.RasterzenSVGOptions
	Timings       bool
	Logger        *log.WOFLogger
}

func NewTileSeeder(w worker.Worker, c cache.Cache) (*TileSeeder, error) {

	logger := log.SimpleWOFLogger()

	s := TileSeeder{
		worker:        w,
		cache:         c,
		SeedRasterzen: true,
		SeedSVG:       true,
		SeedPNG:       false,
		MaxWorkers:    100,
		Timings:       false,
		Logger:        logger,
	}

	return &s, nil
}

func (s *TileSeeder) SeedTileSet(ctx context.Context, ts *TileSet) (bool, []error) {

     golog.Println("START SEEDING")

	t1 := time.Now()

	if s.Timings {

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

	golog.Println("COUNT")
	count := ts.Count()
	golog.Println("COUNT IS", count)

	var remaining int32
	atomic.StoreInt32(&remaining, count)

	if s.Timings {

		ticker := time.NewTicker(time.Second * 5)
		ticker_ch := make(chan bool)

		defer func() {
			ticker_ch <- true
		}()

		go func() {

			for range ticker.C {

				select {
				case <-ticker_ch:
					return
				default:
					r := atomic.LoadInt32(&remaining)
					s.Logger.Status("%d / %d tiles remaining to be processed (%v)", r, count, time.Since(t1))
				}
			}
		}()

	}

	tile_func := func(key, value interface{}) bool {

		t := value.(slippy.Tile)

		go func(t slippy.Tile, throttle chan bool, done_ch chan bool, err_ch chan error) {

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			<-throttle

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			if s.Timings {

				t1 := time.Now()

				defer func() {
					s.Logger.Debug("Time to seed tile (%d/%d/%d) %v", t.Z, t.X, t.Y, time.Since(t1))
				}()
			}

			defer func() {

				k := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)

				err := ts.tile_catalog.Remove(k)

				if err != nil {
					s.Logger.Warning("Failed to remove %s key from tile catalog: %s", k, err)
				}

				done_ch <- true
				throttle <- true
			}()

			ok, errs := s.SeedTiles(t)

			if !ok {

				for _, e := range errs {
					msg := fmt.Sprintf("Unable to seed %v because %s", t, e)
					err_ch <- errors.New(msg)
				}
			}

		}(t, throttle, done_ch, err_ch)

		return true
	}

	ts.Range(tile_func)

	errors := make([]error, 0)

	for atomic.LoadInt32(&remaining) > 0 {
		select {
		case <-done_ch:
			atomic.AddInt32(&remaining, -1)
		case e := <-err_ch:
			errors = append(errors, e)
		default:
			//
		}
	}

	ok := len(errors) == 0
	return ok, errors
}

func (s *TileSeeder) SeedTiles(t slippy.Tile) (bool, []error) {

	if s.SeedRasterzen {

		cache_key := tile.CacheKeyForRasterzenTile(t)
		_, cache_err := s.cache.Get(cache_key)

		if cache_err != nil {

			err := s.worker.RenderRasterzenTile(t)

			if err != nil {
				return false, []error{err}
			}
		}
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)

	remaining := 0

	if s.SeedGeoJSON {

		remaining += 1

		go func() {
			err := s.worker.RenderGeoJSONTile(t)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true
		}()
	}

	if s.SeedExtent {

		remaining += 1

		go func() {
			err := s.worker.RenderExtentTile(t)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true
		}()
	}

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
