package main

import (
	"errors"
	"flag"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"os"
	"strconv"
	"strings"
)

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

	var min_zoom = flag.Int("min-zoom", 1, "The minimum zoom level to fetch for a tile extent.")
	var max_zoom = flag.Int("max-zoom", 16, "The maximum zoom level to fetch for a tile extent.")

	var extents flags.MultiString
	flag.Var(&extents, "extent", "One or more extents to fetch tiles for. Extents should be passed as comma-separated 'minx,miny,maxx,maxy' strings.")

	nextzen_apikey := flag.String("nextzen-apikey", "", "A valid Nextzen API key.")
	nextzen_origin := flag.String("nextzen-origin", "", "An optional HTTP 'Origin' host to pass along with your Nextzen requests.")
	nextzen_debug := flag.Bool("nextzen-debug", false, "Log requests (to STDOUT) to Nextzen tile servers.")
	nextzen_uri := flag.String("nextzen-uri", "", "A valid URI template (RFC 6570) pointing to a custom Nextzen endpoint.")

	var mode = flag.String("mode", "tiles", "Valid modes are: extent, tiles.")

	go_cache := flag.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := flag.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := flag.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	s3_cache := flag.Bool("s3-cache", false, "Cache tiles with a S3-based cache.")
	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	seed_svg := flag.Bool("seed-svg", true, "Seed SVG tiles.")
	seed_png := flag.Bool("seed-png", false, "Seed PNG tiles.")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	writer := io.MultiWriter(os.Stdout)
	logger.AddLogger(writer, "status")

	nz_opts := &nextzen.Options{
		ApiKey: *nextzen_apikey,
		Origin: *nextzen_origin,
		Debug:  *nextzen_debug,
	}

	if *nextzen_uri != "" {

		template, err := uritemplates.Parse(*nextzen_uri)

		if err != nil {
			logger.Fatal(err)
		}

		nz_opts.URITemplate = template
	}

	caches := make([]cache.Cache, 0)

	if *go_cache {

		logger.Info("enable go-cache cache layer")

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			logger.Fatal(err)
		}

		c, err := cache.NewGoCache(opts)

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *fs_cache {

		logger.Info("enable filesystem cache layer")

		if *fs_root == "" {

			cwd, err := os.Getwd()

			if err != nil {
				logger.Fatal(err)
			}

			*fs_root = cwd
		}

		c, err := cache.NewFSCache(*fs_root)

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *s3_cache {

		logger.Info("enable S3 cache layer")

		opts, err := s3.NewS3CacheOptionsFromString(*s3_opts)

		if err != nil {
			logger.Fatal(err)
		}

		c, err := s3.NewS3Cache(*s3_dsn, opts)

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	if len(caches) == 0 {

		// because we still need to pass a cache.Cache thingy
		// around (20180612/thisisaaronland)

		c, err := cache.NewNullCache()

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	c, err := cache.NewMultiCache(caches)

	if err != nil {
		logger.Fatal(err)
	}

	seeder, err := seed.NewTileSeeder(c, nz_opts)

	if err != nil {
		logger.Fatal(err)
	}

	seeder.SeedSVG = *seed_svg
	seeder.SeedPNG = *seed_png
	seeder.Logger = logger

	tileset, err := seed.NewTileSet()

	if err != nil {
		logger.Fatal(err)
	}

	switch strings.ToUpper(*mode) {

	case "EXTENT":

		for _, str_extent := range extents {

			ex, err := parse_extent(str_extent)

			if err != nil {
				logger.Fatal(err)
			}

			for z := *min_zoom; z < *max_zoom; z++ {

				for _, t := range slippy.FromBounds(ex, uint(z)) {
					tileset.AddTile(t)
				}
			}
		}

	case "TILES":

		for _, str_zxy := range flag.Args() {

			z, x, y, err := parse_zxy(str_zxy)

			if err != nil {
				logger.Fatal(err)
			}

			t := slippy.Tile{
				Z: uint(z),
				X: uint(x),
				Y: uint(y),
			}

			tileset.AddTile(t)
		}

	default:
		logger.Fatal("Invalid or unsupported mode")
	}

	ok, errs := seeder.SeedTileSet(tileset)

	if !ok {

		for _, e := range errs {
			logger.Warning(e)
		}

		os.Exit(1)
	}

	os.Exit(0)
}
