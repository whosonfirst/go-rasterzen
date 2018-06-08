package mvt

import (
	"github.com/fapian/geojson2svg/pkg/geojson2svg"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"image"
	"image/png"
	"io"
	"io/ioutil"
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

	for _, l := range layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		features := fc.Get("features")

		for _, f := range features.Array() {

			// https://developer.mozilla.org/en-US/docs/Web/SVG/Tutorial/Fills_and_Strokes
			// https://developer.mozilla.org/en-US/docs/Web/SVG/Tutorial/Paths

			// SOMETHING ABOUT THIS MAKES oksvg VERY VERY SAD BUT I AM NOT SURE WHAT YET...
			// (20180608/thisisaaronland)

			f2, err := sjson.Set(f.String(), "properties.style", "stroke: black; fill: transparent;")

			if err != nil {
				return err
			}

			err = s.AddFeature(f2)

			if err != nil {
				return err
			}
		}
	}

	rsp := s.Draw(256, 256,
		geojson2svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		geojson2svg.WithAttribute("viewBox", "0 0 256 256"),
		geojson2svg.UseProperties([]string{"style", "id", "kind", "kind_detail"}),
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
