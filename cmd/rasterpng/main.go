package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-rasterzen/tile"
	_ "io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	source := flag.String("source", "", "...")
	destination := flag.String("destination", "", "...")	
	
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

		svg_fh, err := os.Open(svg_path)

		if err != nil {
			return err
		}

		defer svg_fh.Close()

		png_path := strings.Replace(svg_path, ext, ".png", 1)
		png_path = strings.Replace(png_path, abs_source, abs_destination, 1)

		log.Println(png_path)
		
		png_root := filepath.Dir(png_path)

		_, err = os.Stat(png_root)

		if os.IsNotExist(err){
			err = os.MkdirAll(png_root, 0755)
		}

		if err != nil {
			return err
		}
		
		// png_fh := ioutil.Discard

		png_fh, err := os.OpenFile(png_path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			return err
		}
		
		err = tile.RasterzenToPNG(svg_fh, png_fh)

		if err != nil {
			return nil
		}
			
		return nil
	}

	cr := crawl.NewCrawler(abs_source)

	err = cr.Crawl(cb)

	if err != nil {
		log.Fatal(err)
	}
}
