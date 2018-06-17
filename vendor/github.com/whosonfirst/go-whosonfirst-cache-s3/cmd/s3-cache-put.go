package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	var dsn = flag.String("dsn", "", "A valid go-whosonfirst-aws DSN string")
	var key = flag.String("key", "", "The name of the key you are setting")
	var value = flag.String("value", "", "The value of the key you are setting. If empty then the code will try to read a file from flag.Args[0]")

	flag.Parse()

	c, err := s3.NewS3Cache(*dsn)

	if err != nil {
		log.Fatal(err)
	}

	if *key == "" {

		args := flag.Args()

		if len(args) == 0 {
			log.Fatal("Missing file to read as key value")
		}

		fh, err := os.Open(args[0])

		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		*value = string(body)
	}

	body, err := cache.SetString(c, *key, "debug")

	if err != nil {

		if cache.IsCacheMiss(err) {
			log.Println("No such key")
			os.Exit(0)
		}

		log.Fatal(err)
	}

	log.Println(body)
}
