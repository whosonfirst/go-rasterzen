package main

import (
	_ "github.com/aaronland/go-cloud-s3blob"
	_ "github.com/whosonfirst/go-cache-blob"
)

import (
	"context"
	"errors"
	"flag"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-cache"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-flags"	
	"io"
	"os"
	"strings"
)

func main() {

	fs := flagset.NewFlagSet("rasterseed")
	
	var min_zoom = fs.Int("min-zoom", 1, "The minimum zoom level to fetch for a tile extent.")
	var max_zoom = fs.Int("max-zoom", 16, "The maximum zoom level to fetch for a tile extent.")

	var extents multi.MultiString
	fs.Var(&extents, "extent", "One or more extents to fetch tiles for. Extents should be passed as comma-separated 'minx,miny,maxx,maxy' strings.")

	nextzen_apikey := fs.String("nextzen-apikey", "", "A valid Nextzen API key.")
	nextzen_origin := fs.String("nextzen-origin", "", "An optional HTTP 'Origin' host to pass along with your Nextzen requests.")
	nextzen_debug := fs.Bool("nextzen-debug", false, "Log requests (to STDOUT) to Nextzen tile servers.")
	nextzen_uri := fs.String("nextzen-uri", "", "A valid URI template (RFC 6570) pointing to a custom Nextzen endpoint.")

	var mode = fs.String("mode", "tiles", "The mode to use when calculating tiles. Valid modes are: extent, tiles.")
	var extent_separator = fs.String("extent-separator", ",", "The separating string for coordinates when calculating tiles in '-mode extent'")

	go_cache := fs.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := fs.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := fs.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	s3_cache := fs.Bool("s3-cache", false, "Cache tiles with a S3-based cache.")
	s3_dsn := fs.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := fs.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	seed_rasterzen := fs.Bool("seed-rasterzen", false, "Seed Rasterzen tiles.")
	seed_svg := fs.Bool("seed-svg", false, "Seed SVG tiles.")
	seed_png := fs.Bool("seed-png", false, "Seed PNG tiles.")
	seed_geojson := fs.Bool("seed-geojson", false, "Seed GeoJSON tiles.")
	seed_extent := fs.Bool("seed-extent", false, "Seed \"extent\" tiles (as GeoJSON).")
	seed_all := fs.Bool("seed-all", false, "See all the tile formats")

	custom_rz_options := fs.String("rasterzen-options", "", "The path to a valid RasterzenOptions JSON file.")
	custom_svg_options := fs.String("svg-options", "", "The path to a valid RasterzenSVGOptions JSON file.")
	custom_png_options := fs.String("png-options", "", "The path to a valid RasterzenPNGOptions JSON file.")

	seed_tileset_catalog_dsn := fs.String("seed-tileset-catalog-dsn", "catalog=sync", "A valid tile.SeedCatalog DSN string. Required parameters are 'catalog=CATALOG'")

	seed_worker := fs.String("seed-worker", "local", "The type of worker for seeding tiles. Valid workers are: lambda, local, sqs.")
	max_workers := fs.Int("seed-max-workers", 100, "The maximum number of concurrent workers to invoke when seeding tiles")

	var lambda_dsn flags.DSNString
	fs.Var(&lambda_dsn, "lambda-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'")

	lambda_function := fs.String("lambda-function", "Rasterzen", "A valid AWS Lambda function name.")

	var sqs_dsn flags.DSNString
	fs.Var(&sqs_dsn, "sqs-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION' and 'queue=QUEUE'")

	timings := fs.Bool("timings", false, "Display timings for tile seeding.")
	strict := fs.Bool("strict", false, "Exit 0 (failure) at the end of seeding a tile set if any errors are encountered.")

	refresh_rasterzen := fs.Bool("refresh-rasterzen", false, "Force rasterzen tiles to be generated even if they are already cached.")
	refresh_svg := fs.Bool("refresh-svg", false, "Force SVG tiles to be generated even if they are already cached.")
	refresh_png := fs.Bool("refresh-png", false, "Force PNG tiles to be generated even if they are already cached.")

	refresh_all := fs.Bool("refresh-all", false, "Force all tiles to be generated even if they are already cached.")

	err := flagset.Parse(fs)

	if err != nil {
		log.Fatal("Failed to parse flagset, %v", err)
	}

	ctx := context.Background()
	
	if *seed_all {
		*seed_rasterzen = true
		*seed_geojson = true
		*seed_extent = true
		*seed_svg = true
		*seed_png = true
	}

	if *refresh_all {
		*refresh_rasterzen = true
		*refresh_svg = true
		*refresh_png = true
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

	for _, u := range cache_uris {

		c, err := cache.NewCache(ctx, u)

		if err != nil {
			log.Fatalf("Failed to instantiate cache '%s', %v", u, err)
		}

		caches = append(caches, c)
	}

	if len(caches) == 0 {

		// because we still need to pass a cache.Cache thingy
		// around (20180612/thisisaaronland)

		c, err := cache.NewCache(ctx, "null://")

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	c, err := cache.NewMultiCache(ctx, caches...)

	if err != nil {
		log.Fatal(err)
	}

	var rz_opts *tile.RasterzenOptions
	var svg_opts *tile.RasterzenSVGOptions
	var png_opts *tile.RasterzenPNGOptions

	if *custom_rz_options != "" {

		var opts *tile.RasterzenOptions

		if strings.HasPrefix(*custom_png_options, "{") {
			opts, err = tile.RasterzenOptionsFromString(*custom_rz_options)
		} else {
			opts, err = tile.RasterzenOptionsFromFile(*custom_rz_options)
		}

		if err != nil {
			logger.Fatal(err)
		}

		rz_opts = opts

	} else {

		opts, err := tile.DefaultRasterzenOptions()

		if err != nil {
			logger.Fatal(err)
		}

		rz_opts = opts
	}

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

		if *custom_png_options != "" {

			var opts *tile.RasterzenPNGOptions

			if strings.HasPrefix(*custom_png_options, "{") {
				opts, err = tile.RasterzenPNGOptionsFromString(*custom_png_options)
			} else {
				opts, err = tile.RasterzenPNGOptionsFromFile(*custom_png_options)
			}

			if err != nil {
				logger.Fatal(err)
			}

			png_opts = opts

		} else {

			opts, err := tile.DefaultRasterzenPNGOptions()

			if err != nil {
				logger.Fatal(err)
			}

			png_opts = opts
		}
	}

	rz_opts.Refresh = *refresh_rasterzen
	svg_opts.Refresh = *refresh_svg
	png_opts.Refresh = *refresh_png

	var w worker.Worker
	var w_err error

	switch strings.ToUpper(*seed_worker) {

	case "LAMBDA":
		w, w_err = worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, c, nz_opts, rz_opts, svg_opts, png_opts)
	case "LOCAL":
		w, w_err = worker.NewLocalWorker(c, nz_opts, rz_opts, svg_opts, png_opts)
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

	_, err = seed.GatherTiles(tileset, seeder, gather_func)

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
