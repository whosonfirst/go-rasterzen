package main

import (
	"flag"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	source := flag.String("source", "", "The path to a directory containing (rasterzen) SVG tiles.")
	destination := flag.String("destination", "", "The path to a directory to write PNG tiles in.")
	dryrun := flag.Bool("dryrun", false, "Go through the motions but don't write anything to disk.")
	force := flag.Bool("force", false, "Re-render already existing PNG files.")

	flag.Parse()

	abs_source, err := filepath.Abs(*source)

	if err != nil {
		log.Fatal(err)
	}

	abs_destination, err := filepath.Abs(*destination)

	if err != nil {
		log.Fatal(err)
	}

	cb := func(svg_path string, info os.FileInfo) error {

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(svg_path)

		if ext != ".svg" {
			return nil
		}

		png_path := strings.Replace(svg_path, ext, ".png", 1)
		png_path = strings.Replace(png_path, abs_source, abs_destination, 1)

		if !*force {
			_, err = os.Stat(png_path)

			if err == nil {
				return nil
			}
		}

		im, err := tile.RasterzenPathToImage(svg_path)

		if err != nil {
			log.Printf("Failed to process %s: %s", svg_path, err)

			if *dryrun {
				return nil
			}

			return err
		}

		var png_fh io.Writer

		if *dryrun {

			png_fh = ioutil.Discard

		} else {

			png_root := filepath.Dir(png_path)

			_, err = os.Stat(png_root)

			if os.IsNotExist(err) {
				err = os.MkdirAll(png_root, 0755)
			}

			if err != nil {
				return err
			}

			fh, err := os.OpenFile(png_path, os.O_RDWR|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			defer fh.Close()
			png_fh = fh
		}

		return png.Encode(png_fh, im)
	}

	cr := crawl.NewCrawler(abs_source)

	err = cr.Crawl(cb)

	if err != nil {
		log.Fatal(err)
	}
}
