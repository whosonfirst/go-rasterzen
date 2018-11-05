package worker

type Options struct {
	SeedSVG bool
	SeedPNG bool
}

type Worker interface {
	SeedTile(slippy.Tile, cache.Cache) error
}
