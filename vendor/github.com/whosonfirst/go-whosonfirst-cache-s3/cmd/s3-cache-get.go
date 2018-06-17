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
	var key = flag.String("key", "", "The name of the key you are setting")

	flag.Parse()

	c, err := s3.NewS3Cache(*dsn)

	if err != nil {
		log.Fatal(err)
	}

	body, err := cache.GetString(c, *key)

	if err != nil {

		if cache.IsCacheMiss(err) {
			log.Println("No such key")
			os.Exit(0)
		}

		log.Fatal(err)
	}

	log.Println(body)
}
