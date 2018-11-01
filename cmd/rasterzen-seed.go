package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// this is basically the http/cache.go GetTileForRequest() function so once we
// have it working here then we should reconcile the two pieces of code...
// (20181101/thisisaaronland)

// something something something what to do about SVG and PNG tiles?
// (20181101/thisisaaronland)

func seed_tile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) error {

	_, err := seed.SeedSVG(t, c, nz_opts)
	return err
}

func parse_zxy(str_zxy string) (int, int, int, error) {

	parts := strings.Split(str_zxy, "/")

	if len(parts) != 3 {
		return 0, 0, 0, errors.New("Invalid xzy string")
	}

	z, err := strconv.Atoi(parts[0])

	if err != nil {
		return 0, 0, 0, err
	}

	x, err := strconv.Atoi(parts[1])

	if err != nil {
		return 0, 0, 0, err
	}

	y, err := strconv.Atoi(parts[2])

	if err != nil {
		return 0, 0, 0, err
	}

	return z, x, y, nil
}

func parse_extent(str_extent string) (*geom.Extent, error) {

	coords := strings.Split(str_extent, " ")

	if len(coords) != 4 {
		return nil, errors.New("Invalid string extent")
	}

	minx, err := strconv.ParseFloat(coords[0], 64)

	if err != nil {
		return nil, err
	}

	miny, err := strconv.ParseFloat(coords[1], 64)

	if err != nil {
		return nil, err
	}

	maxy, err := strconv.ParseFloat(coords[3], 64)

	if err != nil {
		return nil, err
	}

	maxx, err := strconv.ParseFloat(coords[2], 64)

	if err != nil {
		return nil, err
	}

	min := [2]float64{minx, miny}
	max := [2]float64{maxx, maxy}

	return geom.NewExtent(min, max), nil
}

func main() {

	var zoom_levels flags.MultiInt64
	flag.Var(&zoom_levels, "zoom", "...")

	var extents flags.MultiString
	flag.Var(&extents, "extent", "...")

	var api_key = flag.String("nextzen-apikey", "", "...")
	var origin = flag.String("origin", "", "...")

	var mode = flag.String("mode", "tiles", "...")

	go_cache := flag.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := flag.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := flag.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	s3_cache := flag.Bool("s3-cache", false, "Cache tiles with a S3-based cache.")
	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	flag.Parse()

	nz_opts := &nextzen.Options{
		ApiKey: *api_key,
		Origin: *origin,
	}

	caches := make([]cache.Cache, 0)

	if *go_cache {

		log.Println("enable go-cache cache layer")

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			log.Fatal(err)
		}

		c, err := cache.NewGoCache(opts)

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *fs_cache {

		log.Println("enable filesystem cache layer")

		if *fs_root == "" {

			cwd, err := os.Getwd()

			if err != nil {
				log.Fatal(err)
			}

			*fs_root = cwd
		}

		c, err := cache.NewFSCache(*fs_root)

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *s3_cache {

		log.Println("enable S3 cache layer")

		opts, err := s3.NewS3CacheOptionsFromString(*s3_opts)

		if err != nil {
			log.Fatal(err)
		}

		c, err := s3.NewS3Cache(*s3_dsn, opts)

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	if len(caches) == 0 {

		// because we still need to pass a cache.Cache thingy
		// around (20180612/thisisaaronland)

		c, err := cache.NewNullCache()

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	c, err := cache.NewMultiCache(caches)

	if err != nil {
		log.Fatal(err)
	}

	tile_map := new(sync.Map)

	switch strings.ToUpper(*mode) {

	case "EXTENT":

		for _, str_extent := range extents {

			ex, err := parse_extent(str_extent)

			if err != nil {
				log.Fatal(err)
			}

			for _, z := range zoom_levels {

				for _, t := range slippy.FromBounds(ex, uint(z)) {
					k := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
					tile_map.LoadOrStore(k, t)
				}
			}
		}

	case "TILES":

		for _, str_zxy := range flag.Args() {

			z, x, y, err := parse_zxy(str_zxy)

			if err != nil {
				log.Fatal(err)
			}

			t := slippy.Tile{
				Z: uint(z),
				X: uint(x),
				Y: uint(y),
			}

			k := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
			tile_map.LoadOrStore(k, t)
		}

	default:
		log.Fatal("Invalid or unsupported mode")
	}

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

			err := seed_tile(t, c, nz_opts)

			if err != nil {
				msg := fmt.Sprintf("Unabled to seed %v because %s", t, err)
				err_ch <- errors.New(msg)
				return
			}

			log.Println("OK", t)
		}(t)

		return true
	}

	tile_map.Range(tile_func)

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

}
