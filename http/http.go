package http

import (
	"errors"
	"github.com/paulmach/orb/maptile"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"io"
	"log"
	gohttp "net/http"
	"regexp"
	"strconv"
)

var re_path *regexp.Regexp

func init() {
	re_path = regexp.MustCompile(`/(.*)/(\d+)/(\d+)/(\d+).(\w+)$`)
}

func GetTileForRequest(req *gohttp.Request) (io.ReadCloser, error) {

	url := req.URL
	path := url.Path

	if !re_path.MatchString(path) {
		return nil, errors.New("Invalid path")
	}

	m := re_path.FindStringSubmatch(path)

	z, err := strconv.Atoi(m[2])

	if err != nil {
		return nil, err
	}

	x, err := strconv.Atoi(m[3])

	if err != nil {
		return nil, err
	}

	y, err := strconv.Atoi(m[4])

	if err != nil {
		return nil, err
	}

	query := url.Query()
	api_key := query.Get("api_key")

	if api_key == "" {
		return nil, errors.New("Missing API key")
	}

	zm := maptile.Zoom( uint32(z) )
	tl := maptile.Tile{uint32(z), uint32(x), zm}

	log.Println("TILE", tl)
	log.Println("BOUNDS", tl.Bound())

	return nextzen.FetchTile(z, x, y, api_key)
}
