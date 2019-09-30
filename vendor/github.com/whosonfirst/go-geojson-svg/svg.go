// Package geojson2svg provides the SVG type to convert geojson
// geometries, features and featurecollections into a SVG image.
//
// See the tests for usage examples.
package svg

import (
	"bytes"
	"fmt"
	"io"
	_ "log"
	"math"
	"regexp"
	"sort"
	"strings"

	geojson "github.com/paulmach/go.geojson"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/project"
)

// TODO release

type scaleFunc func(float64, float64) (float64, float64)

// SVG represents the SVG that should be created.
// Use the New function to create a SVG. New will handle the defaualt values.
//
// default padding (top: 0, right: 0, bottom: 0, left: 0)
//
// default properties (class)
//
// default attributes ()
type SVG struct {
	useProp            func(string) bool
	padding            Padding
	attributes         map[string]string
	geometries         []*geojson.Geometry
	features           []*geojson.Feature
	featureCollections []*geojson.FeatureCollection
	Mercator           bool
	Extent             *Extent
	Debug              bool
}

// Padding represents the possible padding of the SVG.
type Padding struct{ Top, Right, Bottom, Left float64 }

type Extent struct{ MinX, MinY, MaxX, MaxY float64 }

// An Option represents a single SVG option.
type Option func(*SVG)

// New returns a new SVG that can be used to to draw geojson geometries,
// features and featurecollections.
func New() *SVG {
	return &SVG{
		useProp:    func(prop string) bool { return prop == "class" },
		attributes: make(map[string]string),
		Mercator:   false,
		Extent:     nil,
		Debug:      false,
	}
}

// Draw renders the final SVG with the given options to a string.
// All coordinates will be scaled to fit into the svg.

func (svg *SVG) Draw(width, height float64, opts ...Option) string {

	for _, o := range opts {
		o(svg)
	}

	var sf scaleFunc

	if svg.Mercator && svg.Extent != nil {
		sf = makeScaleFuncMercatorWithExtent(width, height, svg.padding, svg.Extent)
	} else if svg.Mercator {
		sf = makeScaleFuncMercator(width, height, svg.padding, svg.points())
	} else {
		sf = makeScaleFunc(width, height, svg.padding, svg.points())
	}

	content := bytes.NewBufferString("")

	if svg.Debug {
		content.Write([]byte(`<rect width="100%" height="100%" fill="#ff0000" stroke="#000000" stroke-width="1" />`))
	}

	for _, g := range svg.geometries {
		process(sf, content, g, "")
	}
	for _, f := range svg.features {
		as := makeAttributesFromProperties(svg.useProp, f.Properties)
		process(sf, content, f.Geometry, as)
	}
	for _, fc := range svg.featureCollections {
		for _, f := range fc.Features {
			as := makeAttributesFromProperties(svg.useProp, f.Properties)
			process(sf, content, f.Geometry, as)
		}
	}

	attributes := makeAttributes(svg.attributes)
	return fmt.Sprintf(`<svg width="%f" height="%f"%s>%s</svg>`, width, height, attributes, content)
}

// AddGeometry adds a geojson geometry to the svg.
func (svg *SVG) AddGeometry(gs string) error {
	g, err := geojson.UnmarshalGeometry([]byte(gs))
	if err != nil {
		return fmt.Errorf("invalid geometry: %s", gs)
	}
	svg.geometries = append(svg.geometries, g)
	return nil
}

// AddFeature adds a geojson feature to the svg.
func (svg *SVG) AddFeature(fs string) error {
	f, err := geojson.UnmarshalFeature([]byte(fs))
	if err != nil {
		return fmt.Errorf("invalid feature: %s", fs)
	}
	svg.features = append(svg.features, f)
	return nil
}

// AddFeatureCollection adds a geojson featurecollection to the svg.
func (svg *SVG) AddFeatureCollection(fcs string) error {
	fc, err := geojson.UnmarshalFeatureCollection([]byte(fcs))
	if err != nil {
		return fmt.Errorf("invalid feature collection: %s", fcs)
	}
	svg.featureCollections = append(svg.featureCollections, fc)
	return nil
}

// WithAttribute adds the key value pair as attribute to the
// resulting SVG root element.
func WithAttribute(k, v string) Option {
	return func(svg *SVG) {
		svg.attributes[k] = v
	}
}

// WithAttributes adds the map of key value pairs as attributes to the
// resulting SVG root element.
func WithAttributes(as map[string]string) Option {
	return func(svg *SVG) {
		for k, v := range as {
			svg.attributes[k] = v
		}
	}
}

// WithPadding configures the SVG to use the specified padding.
func WithPadding(p Padding) Option {
	return func(svg *SVG) {
		svg.padding = p
	}
}

// UseProperties configures which geojson properties should be copied to the
// resulting SVG element.
func UseProperties(props []string) Option {
	return func(svg *SVG) {
		svg.useProp = func(prop string) bool {
			for _, p := range props {
				if p == prop {
					return true
				}
			}
			return false
		}
	}
}

func (svg *SVG) points() [][]float64 {
	ps := [][]float64{}
	for _, g := range svg.geometries {
		ps = append(ps, collect(g)...)
	}
	for _, f := range svg.features {
		ps = append(ps, collect(f.Geometry)...)
	}
	for _, fs := range svg.featureCollections {
		for _, f := range fs.Features {
			ps = append(ps, collect(f.Geometry)...)
		}
	}
	return ps
}

func process(sf scaleFunc, w io.Writer, g *geojson.Geometry, attributes string) {

	// band-aid: https://github.com/whosonfirst/go-rasterzen/issues/3
	// (20180615/thisisaaronland)

	if g == nil {
		return
	}

	switch {
	case g.IsPoint():
		drawPoint(sf, w, g.Point, attributes)
	case g.IsMultiPoint():
		drawMultiPoint(sf, w, g.MultiPoint, attributes)
	case g.IsLineString():
		drawLineString(sf, w, g.LineString, attributes)
	case g.IsMultiLineString():
		drawMultiLineString(sf, w, g.MultiLineString, attributes)
	case g.IsPolygon():
		drawPolygon(sf, w, g.Polygon, attributes)
	case g.IsMultiPolygon():
		drawMultiPolygon(sf, w, g.MultiPolygon, attributes)
	case g.IsCollection():
		for _, x := range g.Geometries {
			process(sf, w, x, attributes)
		}
	}
}

func collect(g *geojson.Geometry) (ps [][]float64) {

	// band-aid: https://github.com/whosonfirst/go-rasterzen/issues/3
	// (20180615/thisisaaronland)

	if g == nil {
		return
	}

	switch {
	case g.IsPoint():
		ps = append(ps, g.Point)
	case g.IsMultiPoint():
		ps = append(ps, g.MultiPoint...)
	case g.IsLineString():
		ps = append(ps, g.LineString...)
	case g.IsMultiLineString():
		for _, x := range g.MultiLineString {
			ps = append(ps, x...)
		}
	case g.IsPolygon():
		for _, x := range g.Polygon {
			ps = append(ps, x...)
		}
	case g.IsMultiPolygon():
		for _, xs := range g.MultiPolygon {
			for _, x := range xs {
				ps = append(ps, x...)
			}
		}
	case g.IsCollection():
		for _, g := range g.Geometries {
			ps = append(ps, collect(g)...)
		}
	}
	return ps
}

func drawPoint(sf scaleFunc, w io.Writer, p []float64, attributes string) {
	x, y := sf(p[0], p[1])
	fmt.Fprintf(w, `<circle cx="%f" cy="%f" r="1"%s/>`, x, y, attributes)
}

func drawMultiPoint(sf scaleFunc, w io.Writer, ps [][]float64, attributes string) {
	for _, p := range ps {
		drawPoint(sf, w, p, attributes)
	}
}

func drawLineString(sf scaleFunc, w io.Writer, ps [][]float64, attributes string) {
	path := bytes.NewBufferString("M")
	for _, p := range ps {
		x, y := sf(p[0], p[1])
		fmt.Fprintf(path, "%f %f,", x, y)
	}
	fmt.Fprintf(w, `<path d="%s"%s/>`, trim(path), attributes)
}

func drawMultiLineString(sf scaleFunc, w io.Writer, pps [][][]float64, attributes string) {
	for _, ps := range pps {
		drawLineString(sf, w, ps, attributes)
	}
}

func drawPolygon(sf scaleFunc, w io.Writer, pps [][][]float64, attributes string) {
	path := bytes.NewBufferString("")
	for _, ps := range pps {
		subPath := bytes.NewBufferString("M")
		for _, p := range ps {
			x, y := sf(p[0], p[1])
			fmt.Fprintf(subPath, "%f %f,", x, y)
		}
		fmt.Fprintf(path, " %s", trim(subPath))
	}

	fmt.Fprintf(w, `<path d="%s Z"%s/>`, trim(path), attributes)
}

func drawMultiPolygon(sf scaleFunc, w io.Writer, ppps [][][][]float64, attributes string) {
	for _, pps := range ppps {
		drawPolygon(sf, w, pps, attributes)
	}
}

func trim(s fmt.Stringer) string {
	re := regexp.MustCompile(",$")
	return string(re.ReplaceAll([]byte(strings.TrimSpace(s.String())), []byte("")))
}

func makeAttributes(as map[string]string) string {
	keys := make([]string, 0, len(as))
	for k := range as {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	res := bytes.NewBufferString("")
	for _, k := range keys {
		fmt.Fprintf(res, ` %s="%s"`, k, as[k])
	}
	return res.String()
}

func makeAttributesFromProperties(useProp func(string) bool, props map[string]interface{}) string {
	attrs := make(map[string]string)
	for k, v := range props {
		if useProp(k) {
			attrs[k] = fmt.Sprintf("%v", v)
		}
	}
	return makeAttributes(attrs)
}

func makeScaleFunc(width, height float64, padding Padding, ps [][]float64) scaleFunc {

	w := width - padding.Left - padding.Right
	h := width - padding.Top - padding.Bottom

	if len(ps) == 0 {
		return func(x, y float64) (float64, float64) { return x, y }
	}

	if len(ps) == 1 {
		return func(x, y float64) (float64, float64) { return w / 2, h / 2 }
	}

	minX := ps[0][0]
	minY := ps[0][1]

	maxX := ps[0][0]
	maxY := ps[0][1]

	for _, p := range ps[1:] {

		minX = math.Min(minX, p[0])
		minY = math.Min(minY, p[1])

		maxX = math.Max(maxX, p[0])
		maxY = math.Max(maxY, p[1])
	}

	xRes := (maxX - minX) / w
	yRes := (maxY - minY) / h
	res := math.Max(xRes, yRes)

	return func(x, y float64) (float64, float64) {

		x = (x-minX)/res + padding.Left
		y = (maxY-y)/res + padding.Top

		return x, y
	}
}

func makeScaleFuncMercator(width, height float64, padding Padding, ps [][]float64) scaleFunc {

	w := width - padding.Left - padding.Right
	h := height - padding.Top - padding.Bottom

	if len(ps) == 0 {
		return func(x, y float64) (float64, float64) { return x, y }
	}

	if len(ps) == 1 {
		return func(x, y float64) (float64, float64) { return w / 2, h / 2 }
	}

	minX := ps[0][0]
	minY := ps[0][1]

	maxX := ps[0][0]
	maxY := ps[0][1]

	for _, p := range ps[1:] {

		minX = math.Min(minX, p[0])
		minY = math.Min(minY, p[1])

		maxX = math.Max(maxX, p[0])
		maxY = math.Max(maxY, p[1])
	}

	extent := &Extent{
		minX,
		minY,
		maxX,
		maxY,
	}

	return makeScaleFuncMercatorWithExtent(width, height, padding, extent)
}

func makeScaleFuncMercatorWithExtent(width, height float64, padding Padding, extent *Extent) scaleFunc {

	sw_pt := orb.Point{extent.MinX, extent.MinY}
	ne_pt := orb.Point{extent.MaxX, extent.MaxY}

	sw_merc := project.Point(sw_pt, project.WGS84.ToMercator)
	ne_merc := project.Point(ne_pt, project.WGS84.ToMercator)

	minX := sw_merc[0]
	minY := sw_merc[1]

	maxX := ne_merc[0]
	maxY := ne_merc[1]

	w := width - padding.Left - padding.Right
	h := height - padding.Top - padding.Bottom

	xRes := (maxX - minX) / w
	yRes := (maxY - minY) / h

	res := math.Max(xRes, yRes)

	return func(x, y float64) (float64, float64) {

		pt := orb.Point{x, y}
		merc := project.Point(pt, project.WGS84.ToMercator)

		x = merc[0]
		y = merc[1]

		x = (x-minX)/res + padding.Left
		y = (maxY-y)/res + padding.Top

		return x, y
	}
}
