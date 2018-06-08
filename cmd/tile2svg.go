package main

import (
	"flag"
	"github.com/whosonfirst/go-rasterzen/mvt"
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

		err = mvt.ToSVG(fh, out)

		if err != nil {
			log.Fatal(err)
		}

	}
}
