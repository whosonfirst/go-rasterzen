package tile

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	// we are using a forked/patched version - some day
	// hopefully we will just be using this...
	// "github.com/fapian/geojson2svg/pkg/geojson2svg"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/geojson2svg/pkg/geojson2svg"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type SVGStyle struct {
	Stroke        string  `json:"stroke"`
	StrokeWidth   float64 `json:"stroke_width"`
	StrokeOpacity float64 `json:"stroke_opacity"`
	Fill          string  `json:"fill"`
	FillOpacity   float64 `json:"fill_opacity"`
}

type RasterzenSVGStyles map[string]SVGStyle

type RasterzenSVGOptions struct {
	TileSize      float64            `json:"tile_size"`
	Stroke        string             `json:"stroke"`
	StrokeWidth   float64            `json:"stroke_width"`
	StrokeOpacity float64            `json:"stroke_opacity"`
	Fill          string             `json:"fill"`
	FillOpacity   float64            `json:"fill_opacity"`
	DopplrColours bool               `json:"dopplr_colours"`
	Styles        RasterzenSVGStyles `json:"styles"`
}

func DefaultRasterzenSVGOptions() (*RasterzenSVGOptions, error) {

	opts := RasterzenSVGOptions{
		TileSize:      512.0,
		Stroke:        "#000000",
		StrokeWidth:   1.0,
		StrokeOpacity: 1.0,
		Fill:          "#ffffff",
		FillOpacity:   0.5,
		DopplrColours: false,
	}

	return &opts, nil
}

func RasterzenSVGOptionsFromFile(path string) (*RasterzenSVGOptions, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	return RasterzenSVGOptionsFromReader(fh)
}

func RasterzenSVGOptionsFromReader(fh io.Reader) (*RasterzenSVGOptions, error) {

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	return RasterzenSVGOptionsFromBytes(body)
}

func RasterzenSVGOptionsFromBytes(body []byte) (*RasterzenSVGOptions, error) {

	var svg_opts *RasterzenSVGOptions

	err := json.Unmarshal(body, &svg_opts)

	if err != nil {
		return nil, err
	}

	return svg_opts, nil
}

type FeatureCollection struct {
	Type     string        `json:"type"`
	Features []interface{} `json:"features"`
}

func RenderRasterzenTile(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	z := int(t.Z)
	x := int(t.X)
	y := int(t.Y)

	// key := fmt.Sprintf("%d/%d/%d.json", z, x, y)
	// nextzen_key := filepath.Join("nextzen", key)
	// rasterzen_key := filepath.Join("rasterzen", key)

	nextzen_key := CacheKeyForTile(t, "nextzen", "json")
	rasterzen_key := CacheKeyForTile(t, "rasterzen", "json")

	var nextzen_data io.ReadCloser   // stuff sent back from nextzen.org
	var rasterzen_data io.ReadCloser // stuff sent back from nextzen.org

	var err error

	rasterzen_data, err = c.Get(rasterzen_key)

	if err == nil {
		return rasterzen_data, nil
	}

	nextzen_data, err = c.Get(nextzen_key)

	if err != nil {

		t, err := nextzen.FetchTile(z, x, y, nz_opts)

		if err != nil {
			return nil, err
		}

		defer t.Close()

		nextzen_data, err = c.Set(nextzen_key, t)

		if err != nil {
			return nil, err
		}
	}

	cr, err := nextzen.CropTile(z, x, y, nextzen_data)

	if err != nil {
		return nil, err
	}

	defer cr.Close()

	rasterzen_data, err = c.Set(rasterzen_key, cr)

	if err != nil {
		return nil, err
	}

	return rasterzen_data, nil
}

func RasterzenToFeatureCollection(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

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

func RasterzenToGeoJSON(in io.Reader, out io.Writer) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

	type FeatureCollection struct {
		Type     string        `json:"type"`
		Features []interface{} `json:"features"`
	}

	features := make([]interface{}, 0)

	for _, l := range nextzen.Layers {

		fc := gjson.GetBytes(body, l)

		if !fc.Exists() {
			continue
		}

		fc_features := fc.Get("features")

		for _, f := range fc_features.Array() {
			features = append(features, f.Value())
		}
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	enc, err := json.Marshal(fc)

	if err != nil {
		return err
	}

	r := bytes.NewReader(enc)
	_, err = io.Copy(out, r)

	if err != nil {
		return err
	}

	return nil
}

func RasterzenToSVG(in io.Reader, out io.Writer) error {

	opts, err := DefaultRasterzenSVGOptions()

	if err != nil {
		return err
	}

	return RasterzenToSVGWithOptions(in, out, opts)
}

func RasterzenToSVGWithOptions(in io.Reader, out io.Writer, svg_opts *RasterzenSVGOptions) error {

	body, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}

	tile_size := svg_opts.TileSize

	s := geojson2svg.New()
	s.Mercator = true

	use_props := map[string]bool{
		// "id": true,
		"kind":        true,
		"kind_detail": true,
		"sort_rank":   true, // would that SVG had a z-index attribute... applying sort_rank feels like it
		// will be a whole lot of work inside geojson2svg (20180614/thisisaaronland)
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

			stroke := svg_opts.Stroke
			stroke_width := svg_opts.StrokeWidth
			stroke_opacity := svg_opts.StrokeOpacity

			fill := svg_opts.Fill
			fill_opacity := 0.0

			kind := ""
			detail := ""
			geom_type := ""
			sort_rank := int64(0)

			gjson_kind := f.Get("properties.kind")

			if gjson_kind.Exists() {
				kind = gjson_kind.String()
			}

			gjson_detail := f.Get("properties.kind_detail")

			if gjson_detail.Exists() {
				detail = gjson_detail.String()
			}

			gjson_type := f.Get("geometry.type")

			if gjson_type.Exists() {

				geom_type = gjson_type.String()
			}

			gjson_sort := f.Get("properties.sort_rank")

			if gjson_sort.Exists() {
				sort_rank = gjson_sort.Int()
			}

			if geom_type == "Polygon" || geom_type == "MultiPolygon" {

				fill_opacity = svg_opts.FillOpacity
			}

			for query, style := range svg_opts.Styles {

				use_style := true

				for _, str_pair := range strings.Split(query, " ") {

					conditions := strings.Split(str_pair, ".")

					// TO DO: !=
					pair := strings.Split(conditions[1], "=")

					if len(pair) != 2 {
						use_style = false
						break
					}

					k := pair[0]
					v := pair[1]

					possible := make([]string, 0)

					// TO DO: AND queries
					// OR queries

					if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {

						strlen := len(v)

						v = v[1:]
						v = v[:(strlen - 2)]

						possible = strings.Split(v, ",")

					} else {
						possible = []string{v}
					}

					switch conditions[0] {

					case "geometry":

						has_geom := false

						for _, test := range possible {

							if test == geom_type {
								has_geom = true
								break
							}
						}

						use_style = has_geom

					case "properties":

						has_prop := false

						for _, test := range possible {

							path := fmt.Sprintf("properties.%s", k)
							rsp := f.Get(path)

							// log.Println("TEST", path, test)

							if rsp.Exists() && rsp.String() == test {
								has_prop = true
								break
							}

							use_style = has_prop
						}

					default:
						use_style = false
					}
				}

				log.Println("STYLES", query, use_style)

				if !use_style {
					break
				}

				if use_style {

					if style.Stroke != "" {
						stroke = style.Stroke
					}

					if style.StrokeWidth != 0.0 {
						stroke_width = style.StrokeWidth
					}

					if style.StrokeOpacity != 0.0 {
						stroke_opacity = style.StrokeOpacity
					}

					if style.Fill != "" {
						fill = style.Fill
					}

					if style.FillOpacity != 0.0 {
						fill_opacity = style.FillOpacity
					}

					break
				}
			}

			// because we are still working out the details for both
			// logging and sorting features... (20180619/thisisaaronland)

			if false {
				log.Println(kind, detail, geom_type, sort_rank)
			}

			if svg_opts.DopplrColours {
				stroke = str2hex(kind)
				fill = str2hex(detail)
			}

			props := map[string]string{
				"stroke":         stroke,
				"fill":           fill,
				"stroke-width":   strconv.FormatFloat(stroke_width, 'f', -1, 64),
				"stroke-opacity": strconv.FormatFloat(stroke_opacity, 'f', -1, 64),
				"fill-opacity":   strconv.FormatFloat(fill_opacity, 'f', -1, 64),
			}

			for k, v := range props {
				path := fmt.Sprintf("properties.%s", k)
				str_f, _ = sjson.Set(str_f, path, v)
				use_props[k] = true
			}

			// really we should be patching / doing this in geojson2svg
			// but today we're doing it here... (20190529/thisisaaronland)

			if kind != "" {
				kind = escapeXMLString(kind)
				str_f, _ = sjson.Set(str_f, "properties.kind", kind)
			}

			if detail != "" {
				detail = escapeXMLString(detail)
				str_f, _ = sjson.Set(str_f, "properties.kind_detail", detail)
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

func RasterzenToImage(in io.Reader) (image.Image, error) {

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

	err = RasterzenToSVG(in, tmpfile)

	if err != nil {
		return nil, err
	}

	tmpfile.Close()

	return RasterzenPathToImage(tmpfile.Name())
}

func RasterzenPathToImage(path string) (image.Image, error) {

	icon, err := oksvg.ReadIcon(path, oksvg.StrictErrorMode)

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

func RasterzenToPNG(in io.Reader, out io.Writer) error {

	img, err := RasterzenToImage(in)

	if err != nil {
		return err
	}

	return png.Encode(out, img)
}

func str2hex(text string) string {

	hasher := md5.New()
	hasher.Write([]byte(text))

	enc := hex.EncodeToString(hasher.Sum(nil))
	code := enc[0:6]

	return fmt.Sprintf("#%s", code)
}

func escapeXMLString(text string) string {

	// there must be a built-in function to do this...
	// (20190529/thisisaaronland)

	text = strings.Replace(text, "&", "&amp;", -1)
	text = strings.Replace(text, ">", "&gt;", -1)
	text = strings.Replace(text, "<", "&lt;", -1)

	return text
}
