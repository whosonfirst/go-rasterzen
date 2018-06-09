package mvt

import (
	"fmt"
	"github.com/fapian/geojson2svg/pkg/geojson2svg"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func ToSVG(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

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

	s := geojson2svg.New()

	use_props := map[string]bool{
		// "id": true,
		"kind":        true,
		"kind_detail": true,
	}

	for _, l := range layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		features := fc.Get("features")

		for _, f := range features.Array() {

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
				"fill-opacity": "0",
			}

			str_f := f.String()

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

	rsp := s.Draw(256, 256,
		geojson2svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		geojson2svg.WithAttribute("viewBox", "0 0 256 256"),
		geojson2svg.UseProperties(props),
	)

	_, err = out.Write([]byte(rsp))
	return err
}

func ToPNG(in io.Reader, out io.Writer) error {

	tmpfile, err := ioutil.TempFile("", "svg")

	if err != nil {
		return err
	}

	defer func() {

		_, err := os.Stat(tmpfile.Name())

		if !os.IsNotExist(err) {
			os.Remove(tmpfile.Name())
		}
	}()

	err = ToSVG(in, tmpfile)

	if err != nil {
		return err
	}

	tmpfile.Close()

	icon, err := oksvg.ReadIcon(tmpfile.Name(), oksvg.StrictErrorMode)

	if err != nil {
		return err
	}

	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)

	icon.Draw(raster, 1.0)

	return png.Encode(out, img)
}
