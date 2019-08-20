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
	svg_options     *tile.RasterzenSVGOptions
	cache           cache.Cache
	SeedSVG         bool
	SeedPNG         bool
}

func NewLocalWorker(c cache.Cache, nz_opts *nextzen.Options, svg_opts *tile.RasterzenSVGOptions) (Worker, error) {

	w := LocalWorker{
		cache:           c,
		nextzen_options: nz_opts,
		svg_options:     svg_opts,
		SeedSVG:         true,
		SeedPNG:         false,
	}

	return &w, nil
}

func (w *LocalWorker) RenderRasterzenTile(t slippy.Tile) error {

	_, err := tile.RenderRasterzenTile(t, w.cache, w.nextzen_options)
	return err
}

func (w *LocalWorker) RenderGeoJSONTile(t slippy.Tile) error {

	_, err := tile.RenderGeoJSONTile(t, w.cache, w.nextzen_options)
	return err
}

func (w *LocalWorker) RenderExtentTile(t slippy.Tile) error {

	_, err := tile.RenderExtentTile(t, w.cache, w.nextzen_options)
	return err
}

func (w *LocalWorker) RenderSVGTile(t slippy.Tile) error {

	_, err := tile.RenderSVGTile(t, w.cache, w.nextzen_options, w.svg_options)
	return err
}

func (w *LocalWorker) RenderPNGTile(t slippy.Tile) error {

	_, err := tile.RenderPNGTile(t, w.cache, w.nextzen_options, w.svg_options)
	return err
}
