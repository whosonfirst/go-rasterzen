package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"log"
	"os"
)

func main() {

	null_cache := flag.Bool("null-cache", false, "...")
	go_cache := flag.Bool("go-cache", false, "...")
	fs_cache := flag.Bool("fs-cache", false, "...")
	fs_root := flag.String("fs-root", "", "...")

	flag.Parse()

	caches := make([]cache.Cache, 0)

	if *null_cache {

		c, err := cache.NewNullCache()

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *go_cache {

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			log.Fatal(err)
		}

		c, err := cache.NewGoCache(opts)

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *fs_cache {

		if *fs_root == "" {

			cwd, err := os.Getwd()

			if err != nil {
				log.Fatal(err)
			}

			*fs_root = cwd
		}

		c, err := cache.NewFSCache(*fs_root)

		if err != nil {
			log.Fatal(err)
		}

		caches = append(caches, c)
	}

	if len(caches) == 0 {
		log.Fatal("No caches specified")
	}

	c, err := cache.NewMultiCache(caches)

	if err != nil {
		log.Fatal(err)
	}

	args := flag.Args()

	if len(args)%2 == 1 {
		log.Fatal("Arguments not divisible by two (as in 'key -> value')")
	}

	for i := 0; i < len(args); i += 2 {

		k := args[i]
		v := args[i+1]

		g, err := cache.GetString(c, k)

		if err != nil && !cache.IsCacheMiss(err) {
			log.Fatal(err)
		}

		log.Println("GET", k, v, g)

		s, err := cache.SetString(c, k, v)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("SET", k, v, s)

		v2, err := cache.GetString(c, k)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("GET", k, v, v2)
	}

}
