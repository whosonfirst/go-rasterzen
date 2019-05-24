package main

import (
	"flag"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	source := flag.String("source", "", "The path to a directory containing rasterzen (GeoJSON) tiles.")
	destination := flag.String("destination", "", "The path to a directory to SVG tiles in.")
	opts := flag.String("svg-options", "", "The path to a valid RasterzenSVGOptions JSON file.")

	flag.Parse()

	abs_source, err := filepath.Abs(*source)

	if err != nil {
		log.Fatal(err)
	}

	abs_destination, err := filepath.Abs(*destination)

	if err != nil {
		log.Fatal(err)
	}

	svg_opts, err := tile.DefaultRasterzenSVGOptions()

	if err != nil {
		log.Fatal(err)
	}

	if *opts != "" {

		opts, err := tile.RasterzenSVGOptionsFromFile(*opts)

		if err != nil{
			log.Fatal(err)
		}

		svg_opts = opts
	}
	
	cb := func(rasterzen_path string, info os.FileInfo) error {

		if info.IsDir() {
			return nil
		}

		log.Println(rasterzen_path)
		
		ext := filepath.Ext(rasterzen_path)

		if ext != ".json" {
			return nil
		}

		rasterzen_fh, err := os.Open(rasterzen_path)

		if err != nil {
			return err
		}

		svg_path := strings.Replace(rasterzen_path, ext, ".svg", 1)
		svg_path = strings.Replace(svg_path, abs_source, abs_destination, 1)

		log.Println(svg_path)

		svg_root := filepath.Dir(svg_path)

		_, err = os.Stat(svg_root)

		if os.IsNotExist(err) {
			err = os.MkdirAll(svg_root, 0755)
		}

		if err != nil {
			return err
		}

		// svg_fh := ioutil.Discard

		svg_fh, err := os.OpenFile(svg_path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			return err
		}

		defer svg_fh.Close()
		
		err = tile.RasterzenToSVGWithOptions(rasterzen_fh, svg_fh, svg_opts)

		if err != nil {
			return err
		}

		return nil
	}

	cr := crawl.NewCrawler(abs_source)

	err = cr.Crawl(cb)

	if err != nil {
		log.Fatal(err)
	}
}
