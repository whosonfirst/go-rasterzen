package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"log"
	"os"
)

func main() {

	var dsn = flag.String("dsn", "", "A valid go-whosonfirst-aws DSN string")
	var opts = flag.String("opts", "", "...")
	var key = flag.String("key", "", "The name of the key you are setting")

	flag.Parse()

	s3_opts, err := s3.NewS3CacheOptionsFromString(*opts)

	if err != nil {
		log.Fatal(err)
	}

	s3_cache, err := s3.NewS3Cache(*dsn, s3_opts)

	if err != nil {
		log.Fatal(err)
	}

	body, err := cache.GetString(s3_cache, *key)

	if err != nil {

		if cache.IsCacheMiss(err) {
			log.Println("No such key")
			os.Exit(0)
		}

		log.Fatal(err)
	}

	log.Println(body)
}
