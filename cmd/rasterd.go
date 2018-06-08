package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-rasterzen/http"
	"log"
	gohttp "net/http"
	"os"
)

func main() {

	var host = flag.String("host", "localhost", "The host for rasterd to listen for requests on")
	var port = flag.Int("port", 8080, "The port for rasterd to listen for requests on")

	flag.Parse()

	mux := gohttp.NewServeMux()

	png_handler, err := http.PNGHandler()

	if err != nil {
		log.Fatal(err)
	}

	svg_handler, err := http.SVGHandler()

	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("/png/", png_handler)
	mux.Handle("/svg/", svg_handler)

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("listening for requests on %s\n", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
