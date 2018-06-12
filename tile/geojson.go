package tile

import (
	"encoding/json"
	"fmt"
	"github.com/fapian/geojson2svg/pkg/geojson2svg"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	_ "log"
	"os"
)

type FeatureCollection struct {
	Type     string        `json:"type"`
	Features []interface{} `json:"features"`
}

func ToFeatureCollection(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

	/*
		layers := []string{
			"boundaries",
			"buildings",
			"earth",
			"landuse",
			"places",
			"pois",
			"roads",
			"transit",
			"water",
		}
	*/

	features := make([]interface{}, 0)

	for _, l := range nextzen.Layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		gjson_features := fc.Get("features")

		for _, f := range gjson_features.Array() {
			features = append(features, f.Value())
		}
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	b, err := json.Marshal(fc)

	if err != nil {
		return err
	}

	_, err = out.Write(b)
	return err
}

func ToSVG(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

	tile_size := 512.0

	s := geojson2svg.New()
	s.Mercator = true

	use_props := map[string]bool{
		// "id": true,
		"kind":        true,
		"kind_detail": true,
	}

	for _, l := range nextzen.Layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		features := fc.Get("features")

		for _, f := range features.Array() {

			str_f := f.String()

			// https://developer.mozilla.org/en-US/docs/Web/SVG/Tutorial/Fills_and_Strokes
			// https://developer.mozilla.org/en-US/docs/Web/SVG/Tutorial/Paths

			// oksvg does not know how to deal with named style properties (like transparent)
			// https://github.com/srwiley/oksvg/blob/master/doc/SVG_Element_List.txt

			// we could also define a single "style" attribute but this just feels a bit
			// more explicit and easier to compose on a per-layer basis in the future
			// (20180608/thisisaaronland)

			props := map[string]string{
				"stroke":       "#000000",
				"fill":         "#ffffff", // "transparent",
				"fill-opacity": "0.25",
			}

			for k, v := range props {
				path := fmt.Sprintf("properties.%s", k)
				str_f, _ = sjson.Set(str_f, path, v)
				use_props[k] = true
			}

			err := s.AddFeature(str_f)

			if err != nil {
				return err
			}
		}
	}

	props := make([]string, 0)

	for k, _ := range use_props {
		props = append(props, k)
	}

	rsp := s.Draw(tile_size, tile_size,
		geojson2svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		geojson2svg.WithAttribute("viewBox", fmt.Sprintf("0 0 %d %d", int(tile_size), int(tile_size))),
		geojson2svg.UseProperties(props),
	)

	_, err = out.Write([]byte(rsp))
	return err
}

func ToImage(in io.Reader) (image.Image, error) {

	tmpfile, err := ioutil.TempFile("", "svg")

	if err != nil {
		return nil, err
	}

	defer func() {

		_, err := os.Stat(tmpfile.Name())

		if !os.IsNotExist(err) {
			os.Remove(tmpfile.Name())
		}
	}()

	err = ToSVG(in, tmpfile)

	if err != nil {
		return nil, err
	}

	tmpfile.Close()

	icon, err := oksvg.ReadIcon(tmpfile.Name(), oksvg.StrictErrorMode)

	if err != nil {
		return nil, err
	}

	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)

	icon.Draw(raster, 1.0)

	return img, nil
}

func ToPNG(in io.Reader, out io.Writer) error {

	img, err := ToImage(in)

	if err != nil {
		return err
	}

	return png.Encode(out, img)
}
