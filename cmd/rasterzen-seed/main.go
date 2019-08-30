package main

import (
	"context"
	"errors"
	"flag"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"os"
	"strings"
)

func main() {

	var min_zoom = flag.Int("min-zoom", 1, "The minimum zoom level to fetch for a tile extent.")
	var max_zoom = flag.Int("max-zoom", 16, "The maximum zoom level to fetch for a tile extent.")

	var extents flags.MultiString
	flag.Var(&extents, "extent", "One or more extents to fetch tiles for. Extents should be passed as comma-separated 'minx,miny,maxx,maxy' strings.")

	nextzen_apikey := flag.String("nextzen-apikey", "", "A valid Nextzen API key.")
	nextzen_origin := flag.String("nextzen-origin", "", "An optional HTTP 'Origin' host to pass along with your Nextzen requests.")
	nextzen_debug := flag.Bool("nextzen-debug", false, "Log requests (to STDOUT) to Nextzen tile servers.")
	nextzen_uri := flag.String("nextzen-uri", "", "A valid URI template (RFC 6570) pointing to a custom Nextzen endpoint.")

	var mode = flag.String("mode", "tiles", "The mode to use when calculating tiles. Valid modes are: extent, tiles.")
	var extent_separator = flag.String("extent-separator", ",", "The separating string for coordinates when calculating tiles in '-mode extent'")

	go_cache := flag.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := flag.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := flag.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	s3_cache := flag.Bool("s3-cache", false, "Cache tiles with a S3-based cache.")
	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	seed_rasterzen := flag.Bool("seed-rasterzen", false, "Seed Rasterzen tiles.")
	seed_svg := flag.Bool("seed-svg", false, "Seed SVG tiles.")
	seed_png := flag.Bool("seed-png", false, "Seed PNG tiles.")
	seed_geojson := flag.Bool("seed-geojson", false, "Seed GeoJSON tiles.")
	seed_extent := flag.Bool("seed-extent", false, "Seed \"extent\" tiles (as GeoJSON).")
	seed_all := flag.Bool("seed-all", false, "See all the tile formats")

	custom_svg_options := flag.String("svg-options", "", "The path to a valid RasterzenSVGOptions JSON file.")

	seed_tileset_catalog_dsn := flag.String("seed-tileset-catalog-dsn", "catalog=sync", "A valid tile.SeedCatalog DSN string. Required parameters are 'catalog=CATALOG'")

	seed_worker := flag.String("seed-worker", "local", "The type of worker for seeding tiles. Valid workers are: lambda, local, sqs.")
	max_workers := flag.Int("seed-max-workers", 100, "The maximum number of concurrent workers to invoke when seeding tiles")

	var lambda_dsn flags.DSNString
	flag.Var(&lambda_dsn, "lambda-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'")

	lambda_function := flag.String("lambda-function", "Rasterzen", "A valid AWS Lambda function name.")

	var sqs_dsn flags.DSNString
	flag.Var(&sqs_dsn, "sqs-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION' and 'queue=QUEUE'")

	timings := flag.Bool("timings", false, "Display timings for tile seeding.")
	strict := flag.Bool("strict", false, "Exit 0 (failure) at the end of seeding a tile set if any errors are encountered.")

	flag.Parse()

	if *seed_all {
		*seed_rasterzen = true
		*seed_geojson = true
		*seed_extent = true
		*seed_svg = true
		*seed_png = true
	}

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

	var svg_opts *tile.RasterzenSVGOptions

	if *seed_svg || *seed_png {

		if *custom_svg_options != "" {

			var opts *tile.RasterzenSVGOptions

			if strings.HasPrefix(*custom_svg_options, "{") {
				opts, err = tile.RasterzenSVGOptionsFromString(*custom_svg_options)
			} else {
				opts, err = tile.RasterzenSVGOptionsFromFile(*custom_svg_options)
			}

			if err != nil {
				logger.Fatal(err)
			}

			svg_opts = opts

		} else {

			opts, err := tile.DefaultRasterzenSVGOptions()

			if err != nil {
				logger.Fatal(err)
			}

			svg_opts = opts
		}

	}

	var w worker.Worker
	var w_err error

	switch strings.ToUpper(*seed_worker) {

	case "LAMBDA":
		w, w_err = worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, c, nz_opts, svg_opts)
	case "LOCAL":
		w, w_err = worker.NewLocalWorker(c, nz_opts, svg_opts)
	case "SQS":
		w, w_err = worker.NewSQSWorker(sqs_dsn.Map())
	default:
		w_err = errors.New("Invalid worker")

	}

	if w_err != nil {
		logger.Fatal(w_err)
	}

	seeder, err := seed.NewTileSeeder(w, c)

	if err != nil {
		logger.Fatal(err)
	}

	seeder.MaxWorkers = *max_workers
	seeder.Logger = logger
	seeder.Timings = *timings

	seeder.SeedRasterzen = *seed_rasterzen
	seeder.SeedSVG = *seed_svg
	seeder.SeedPNG = *seed_png
	seeder.SeedGeoJSON = *seed_geojson
	seeder.SeedExtent = *seed_extent

	tileset, err := seed.NewTileSetFromDSN(*seed_tileset_catalog_dsn)

	if err != nil {
		logger.Fatal(err)
	}

	tileset.Logger = logger
	tileset.Timings = *timings

	var gather_func seed.GatherTilesFunc

	switch strings.ToUpper(*mode) {

	case "EXTENT":
		gather_func, err = seed.NewGatherTilesExtentFunc(extents, *extent_separator, *min_zoom, *max_zoom)
	case "TILES":
		gather_func, err = seed.NewGatherTilesFunc(flag.Args())
	default:
		err = errors.New("Invalid or unsupported mode")
	}

	if err != nil {
		logger.Fatal(err)
	}

	err = seed.GatherTiles(tileset, seeder, gather_func)

	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ok, errs := seeder.SeedTileSet(ctx, tileset)

	if !ok {

		for _, e := range errs {
			logger.Warning(e)
		}

		if *strict {
			os.Exit(1)
		}
	}

	os.Exit(0)
}
