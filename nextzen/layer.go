package nextzen

var Layers []string

func init() {

	// things that are disabled are pending work to wrangle sort rank
	// and not drawing (or rather filling) things that draw grid lines
	// careful readers will note that water is still enabled but that's
	// because it makes for "tasteful" grid lines... I hate computers
	// (20180615/thisisaaronland)

	Layers = []string{
		// "boundaries",
		"buildings",
		//"earth",
		// "landuse",
		"places",
		"pois",
		"roads",
		"transit",
		"water",
	}
}
