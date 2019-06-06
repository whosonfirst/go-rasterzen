package nextzen

// should this be in tile/nextgen.go ? perhaps...
// (20190606/thisisaaronland)

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/jtacoma/uritemplates"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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

func FetchTile(z int, x int, y int, opts *Options) (io.ReadCloser, error) {

	fetch_z := z
	fetch_x := x
	fetch_y := y

	overzoom := false

	if z > 16 {
		overzoom = true
	}

	if overzoom {

		u16 := uint(16)
		uz := uint(z)
		ux := uint(x)
		uy := uint(y)

		mag := uz - u16
		t := slippy.NewTile(u16, ux>>mag, uy>>mag)

		fetch_z = int(t.Z)
		fetch_x = int(t.X)
		fetch_y = int(t.Y)
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

	if overzoom {

		body, err := ioutil.ReadAll(rsp_body)

		if err != nil {
			return nil, err
		}

		// CROP ZOOM 16 TILE HERE YEAH...

		cropped_doc := make(map[string]interface{})

		type CroppedResponse struct {
			Layer string
			Body  interface{}
		}

		done_ch := make(chan bool)
		err_ch := make(chan error)
		rsp_ch := make(chan CroppedResponse)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for _, l := range Layers {

			go func(l string) {

				defer func() {
					done_ch <- true
				}()

				select {
				case <-ctx.Done():
					return
				default:
					// pass
				}

				fc := gjson.GetBytes(body, l)

				if !fc.Exists() {
					return
				}

				enc, err := json.Marshal(fc.Value())

				if err != nil {
					err_ch <- err
					return
				}

				r := bytes.NewReader(enc)
				fh := ioutil.NopCloser(r)

				cropped, err := CropTile(z, x, y, fh)

				if err != nil {
					err_ch <- err
					return
				}

				cropped_body, err := ioutil.ReadAll(cropped)

				if err != nil {
					err_ch <- err
					return
				}

				var stub interface{}

				err = json.Unmarshal(cropped_body, &stub)

				if err != nil {
					err_ch <- err
					return
				}

				rsp := CroppedResponse{
					Layer: l,
					Body:  stub,
				}

				rsp_ch <- rsp
			}(l)
		}

		remaining := len(Layers)

		for remaining > 0 {

			select {
			case <-done_ch:
				// log.Println("DONE")
				remaining -= 1
			case err := <-err_ch:
				// log.Println("OH NO", err)
				return nil, err
			case rsp := <-rsp_ch:
				// log.Println("CROPPED", rsp.Layer)
				cropped_doc[rsp.Layer] = rsp.Body
			}
		}

		enc, err := json.Marshal(cropped_doc)

		if err != nil {
			return nil, err
		}

		log.Println("CROPPED", string(enc))

		r := bytes.NewReader(enc)
		rsp_body = ioutil.NopCloser(r)
	}

	return rsp_body, nil
}

func CropTile(z int, x int, y int, fh io.ReadCloser) (io.ReadCloser, error) {

	zm := maptile.Zoom(uint32(z))
	tl := maptile.New(uint32(x), uint32(y), zm)

	bounds := tl.Bound()

	// log.Println("CROP", z, x, y, bounds.Min.Lon(), bounds.Min.Lat(), bounds.Max.Lon(), bounds.Max.Lat())

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
	return ioutil.NopCloser(r), nil
}
