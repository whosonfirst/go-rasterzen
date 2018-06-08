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

	var api_key = flag.String("nextzen-apikey", "", "A valid Nextzen API key (https://developers.nextzen.org/)")

	flag.Parse()

	mux := gohttp.NewServeMux()

	svg_handler, err := http.SVGHandler(*api_key)

	mux.Handle("/", svg_handler)

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("listening for requests on %s\n", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
