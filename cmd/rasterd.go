package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-rasterzen/http"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"log"
	gohttp "net/http"
	"os"
)

func main() {

	var host = flag.String("host", "localhost", "The host for rasterd to listen for requests on.")
	var port = flag.Int("port", 8080, "The port for rasterd to listen for requests on.")

	no_cache := flag.Bool("no-cache", false, "Disable all caching.")
	go_cache := flag.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := flag.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := flag.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	// fs_ttl := flag.Int("fs-ttl", 0, "The time-to-live (in seconds) for filesystem cache files. If 0 cached tiles will never expire.")

	png_handler := flag.Bool("png-handler", true, "Enable the PNG tile handler.")
	svg_handler := flag.Bool("svg-handler", true, "Enable the SVG tile handler.")
	geojson_handler := flag.Bool("geojson-handler", false, "Enable the GeoJSON tile handler.")

	flag.Parse()

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

	mux := gohttp.NewServeMux()

	if *png_handler {

		log.Println("enable PNG handler")

		h, err := http.PNGHandler(c)

		if err != nil {
			log.Fatal(err)
		}

		mux.Handle("/png/", h)
	}

	if *svg_handler {

		log.Println("enable SVG handler")

		h, err := http.SVGHandler(c)

		if err != nil {
			log.Fatal(err)
		}

		mux.Handle("/svg/", h)
	}

	if *geojson_handler {

		log.Println("enable GeoJSON handler")

		h, err := http.GeoJSONHandler(c)

		if err != nil {
			log.Fatal(err)
		}

		mux.Handle("/geojson/", h)
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("listening for requests on %s\n", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
