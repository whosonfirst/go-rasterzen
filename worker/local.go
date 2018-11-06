package worker

import (
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
)

type LocalWorker struct {
	Worker
	nextzen_options *nextzen.Options
	cache           cache.Cache
	SeedSVG         bool
	SeedPNG         bool
}

func NewLocalWorker(c cache.Cache, nz_opts *nextzen.Options) (Worker, error) {

	w := LocalWorker{
		cache:           c,
		nextzen_options: nz_opts,
		SeedSVG:         true,
		SeedPNG:         false,
	}

	return &w, nil
}

func (w *LocalWorker) RenderRasterzenTile(t slippy.Tile) error {

	_, err := tile.RenderRasterzenTile(t, w.cache, w.nextzen_options)
	return err
}

func (w *LocalWorker) RenderSVGTile(t slippy.Tile) error {

	_, err := tile.RenderSVGTile(t, w.cache, w.nextzen_options)
	return err
}

func (w *LocalWorker) RenderPNGTile(t slippy.Tile) error {

	_, err := tile.RenderPNGTile(t, w.cache, w.nextzen_options)
	return err
}
