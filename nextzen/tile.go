package nextzen

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"log"
	_ "net/http"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func FetchTile(z int, x int, y int, api_key string) (io.ReadCloser, error) {

	layer := "all"

	url := fmt.Sprintf("https://tile.nextzen.org/tilezen/vector/v1/256/%s/%d/%d/%d.json?api_key=%s", layer, z, x, y, api_key)

	rsp, err := http.Get(url)

	/*
		cl, err := NewHTTPClient()

		if err != nil {
			return nil, err
		}

		rsp, err := cl.Get(url)
	*/

	if err != nil {
		return nil, err
	}

	// for reasons I don't understand the following does not appear
	// to trigger an error (20180628/thisisaaronland)
	// < HTTP/2 400
	// < content-length: 16
	// < server: CloudFront
	// < date: Thu, 28 Jun 2018 20:50:44 GMT
	// < age: 73
	// < x-cache: Error from cloudfront
	// < via: 1.1 02192a27c967e955f8c815efa939bfc8.cloudfront.net (CloudFront)
	// < x-amz-cf-id: m42n6AwT9N-kBNzKnrKxe1eXfQITapw0BAfE8kG89vPNn0rQ2TQKTg==

	if rsp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Nextzen returned a non-200 response fetching %s/%d/%d/%d : '%s'", layer, z, x, y, rsp.Status))
	}

	return rsp.Body, nil
}

func CropTile(z int, x int, y int, fh io.ReadCloser) (io.ReadCloser, error) {

	zm := maptile.Zoom(uint32(z))
	tl := maptile.Tile{uint32(x), uint32(y), zm}

	bounds := tl.Bound()

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	// Layers is defined in nextzen/layers.go

	for _, l := range Layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		features := fc.Get("features")

		// SOMETHING SOMETHING SOMETHING DO THESE ALL IN PARALLEL...

		for i, f := range features.Array() {

			str_f := f.String()

			feature, err := geojson.UnmarshalFeature([]byte(str_f))

			if err != nil {
				return nil, err
			}

			geom := feature.Geometry

			orb_geom := clip.Geometry(bounds, geom)
			new_geom := geojson.NewGeometry(orb_geom)

			path := fmt.Sprintf("%s.features.%d.geometry", l, i)
			body, err = sjson.SetBytes(body, path, new_geom)

			if err != nil {
				return nil, err
			}
		}
	}

	r := bytes.NewReader(body)
	return nopCloser{r}, nil
}
