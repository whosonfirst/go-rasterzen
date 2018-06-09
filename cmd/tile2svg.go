package main

import (
	"flag"
	"github.com/whosonfirst/go-rasterzen/tile"
	"log"
	"os"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		// log.Println(path)

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		defer fh.Close()

		out := os.Stdout

		err = tile.ToSVG(fh, out)

		if err != nil {
			log.Fatal(err)
		}

	}
}
