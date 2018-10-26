package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-rasterzen/http"
	"github.com/whosonfirst/go-rasterzen/server"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"log"
	gohttp "net/http"
	gourl "net/url"
	"os"
)

func main() {

	config := flag.String("config", "", "Read some or all flags from an ini-style config file. Values in the config file take precedence over command line flags.")
	section := flag.String("section", "rasterd", "A valid ini-style config file section.")

	var proto = flag.String("protocol", "http", "The protocol for wof-staticd server to listen on. Valid protocols are: http, lambda.")
	var host = flag.String("host", "localhost", "The host for rasterd to listen for requests on.")
	var port = flag.Int("port", 8080, "The port for rasterd to listen for requests on.")

	no_cache := flag.Bool("no-cache", false, "Disable all caching.")
	go_cache := flag.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := flag.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := flag.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	s3_cache := flag.Bool("s3-cache", false, "Cache tiles with a S3-based cache.")
	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	nextzen_apikey := flag.String("nextzen-apikey", "", "...")
	nextzen_origin := flag.String("nextzen-origin", "", "...")
	nextzen_debug := flag.Bool("nextzen-debug", false, "...")

	// fs_ttl := flag.Int("fs-ttl", 0, "The time-to-live (in seconds) for filesystem cache files. If 0 cached tiles will never expire.")

	png_handler := flag.Bool("png-handler", true, "Enable the PNG tile handler.")
	svg_handler := flag.Bool("svg-handler", true, "Enable the SVG tile handler.")
	geojson_handler := flag.Bool("geojson-handler", false, "Enable the GeoJSON tile handler.")

	var path_png = flag.String("path-png", "/png/", "The path that PNG tiles should be served from")
	var path_svg = flag.String("path-svg", "/svg/", "The path that SVG tiles should be served from")
	var path_geojson = flag.String("path-geojson", "/geojson/", "The path that GeoJSON tiles should be served from")

	flag.Parse()

	if *config != "" {

		err := flags.SetFlagsFromConfig(*config, *section)

		if err != nil {
			log.Fatal(err)
		}

	} else {

		err := flags.SetFlagsFromEnvVars("RASTERD")

		if err != nil {
			log.Fatal(err)
		}
	}

	if *no_cache {

		log.Println("disable all cache layers")

		*go_cache = false
		*fs_cache = false
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

	ch, err := http.NewCacheHandler(c)

	if err != nil {
		log.Fatal(err)
	}

	if *nextzen_apikey != "" {
		ch.NextzenOptions.ApiKey = *nextzen_apikey
	}

	if *nextzen_origin != "" {
		ch.NextzenOptions.Origin = *nextzen_origin
	}

	ch.NextzenOptions.Debug = *nextzen_debug

	mux := gohttp.NewServeMux()

	if *png_handler {

		log.Println("enable PNG handler")

		h, err := http.PNGHandler(ch)

		if err != nil {
			log.Fatal(err)
		}

		mux.Handle(*path_png, h)
	}

	if *svg_handler {

		log.Println("enable SVG handler")

		h, err := http.SVGHandler(ch)

		if err != nil {
			log.Fatal(err)
		}

		mux.Handle(*path_svg, h)
	}

	if *geojson_handler {

		log.Println("enable GeoJSON handler")

		h, err := http.GeoJSONHandler(ch)

		if err != nil {
			log.Fatal(err)
		}

		mux.Handle(*path_geojson, h)
	}

	address := fmt.Sprintf("http://%s:%d", *host, *port)

	u, err := gourl.Parse(address)

	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewStaticServer(*proto, u)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s\n", s.Address())

	err = s.ListenAndServe(mux)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
