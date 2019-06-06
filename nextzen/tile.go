package nextzen

// should this be in tile/nextgen.go ? perhaps...
// (20190606/thisisaaronland)

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jtacoma/uritemplates"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

type Options struct {
	ApiKey      string
	Origin      string
	Debug       bool
	URITemplate *uritemplates.UriTemplate
}

var default_endpoint *uritemplates.UriTemplate

func init() {

	template := "https://tile.nextzen.org/tilezen/vector/v1/256/{layer}/{z}/{x}/{y}.json?api_key={apikey}"
	default_endpoint, _ = uritemplates.Parse(template)
}

func MaxZoom() int {
	return 16
}

func IsOverZoom(z int) bool {

	if z > MaxZoom() {
		return true
	}

	return false
}

func FetchTile(z int, x int, y int, opts *Options) (io.ReadCloser, error) {

	fetch_z := z
	fetch_x := x
	fetch_y := y

	// see notes below about whether or not we keep the overzooming code
	// in this package or in tile/rasterzen.go (20190606/thisisaaronland)

	overzoom := IsOverZoom(z)

	if overzoom {

		max := MaxZoom()
		mag := z - max

		ux := uint(x) >> uint(mag)
		uy := uint(y) >> uint(mag)

		fetch_z = max
		fetch_x = int(ux)
		fetch_y = int(uy)
	}

	layer := "all"

	values := make(map[string]interface{})
	values["layer"] = "all"
	values["apikey"] = opts.ApiKey
	values["z"] = fetch_z
	values["x"] = fetch_x
	values["y"] = fetch_y

	endpoint := default_endpoint

	if opts.URITemplate != nil {
		endpoint = opts.URITemplate
	}

	url, err := endpoint.Expand(values)

	if err != nil {
		return nil, err
	}

	cl := new(http.Client)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	if opts.Origin != "" {
		req.Header.Set("Origin", opts.Origin)
	}

	if opts.Debug {

		dump, err := httputil.DumpRequest(req, false)

		if err != nil {
			return nil, err
		}

		log.Println(string(dump))
	}

	rsp, err := cl.Do(req)

	if err != nil {
		return nil, err
	}

	if opts.Debug {
		log.Println(url, rsp.Status)
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

	rsp_body := rsp.Body

	// overzooming works until it doesn't - specifically there are
	// weird offsets that I don't understand - examples include:
	// ./bin/rasterd -www -no-cache -nextzen-debug -nextzen-apikey {KEY}
	// http://localhost:8080/#18/37.61800/-122.38301
	// http://localhost:8080/#19/37.61780/-122.38800
	// http://localhost:8080/svg/19/83903/202936.svg?api_key={KEY}
	// (20190606/thisisaaronland)

	if overzoom {

		// it would be good to cache rsp_body (aka the Z16 tile) here or maybe
		// we just move all of this logic in to tile/rasterzen.go...
		// (20190606/thisisaaronland)

		cropped_rsp, err := CropTile(z, x, y, rsp_body)

		if err != nil {
			return nil, err
		}

		rsp_body = cropped_rsp
	}

	return rsp_body, nil
}

// crop all the elements in fh to the bounds of (z, x, y)

func CropTile(z int, x int, y int, fh io.ReadCloser) (io.ReadCloser, error) {

	zm := maptile.Zoom(uint32(z))
	tl := maptile.New(uint32(x), uint32(y), zm)

	bounds := tl.Bound()

	// log.Println("CROP", z, x, y, bounds.Min.Lon(), bounds.Min.Lat(), bounds.Max.Lon(), bounds.Max.Lat())

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	cropped_tile := make(map[string]interface{})

	type CroppedResponse struct {
		Layer             string
		FeatureCollection *geojson.FeatureCollection
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)
	rsp_ch := make(chan CroppedResponse)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Layers is defined in nextzen/layers.go

	for _, layer_name := range Layers {

		go func(layer_name string) {

			defer func() {
				done_ch <- true
			}()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			fc_rsp := gjson.GetBytes(body, layer_name)

			if !fc_rsp.Exists() {
				return
			}

			fc_str := fc_rsp.String()

			fc, err := geojson.UnmarshalFeatureCollection([]byte(fc_str))

			if err != nil {
				err_ch <- err
				return
			}

			cropped_fc := geojson.NewFeatureCollection()

			for _, f := range fc.Features {

				geom := f.Geometry
				clipped_geom := clip.Geometry(bounds, geom)

				// I wish clip.Geometry returned errors rather than
				// silently not clipping anything...
				// https://github.com/paulmach/orb/blob/master/clip/helpers.go#L11-L23

				if clipped_geom == nil {
					continue
				}

				f.Geometry = clipped_geom
				cropped_fc.Append(f)
			}

			rsp := CroppedResponse{
				Layer:             layer_name,
				FeatureCollection: cropped_fc,
			}

			rsp_ch <- rsp

		}(layer_name)
	}

	remaining := len(Layers)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, err
		case rsp := <-rsp_ch:
			cropped_tile[rsp.Layer] = rsp.FeatureCollection
		}
	}

	cropped_body, err := json.Marshal(cropped_tile)

	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(cropped_body)
	return ioutil.NopCloser(r), nil
}
