package worker

import (
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
)

type LocalWorker struct {
	Worker
	nextzen_options   *nextzen.Options
	rasterzen_options *tile.RasterzenOptions
	svg_options       *tile.RasterzenSVGOptions
	png_options       *tile.RasterzenPNGOptions
	cache             cache.Cache
	SeedSVG           bool
	SeedPNG           bool
}

func NewLocalWorker(c cache.Cache, nz_opts *nextzen.Options, rz_opts *tile.RasterzenOptions, svg_opts *tile.RasterzenSVGOptions, png_opts *tile.RasterzenPNGOptions) (Worker, error) {

	w := LocalWorker{
		cache:             c,
		nextzen_options:   nz_opts,
		rasterzen_options: rz_opts,
		svg_options:       svg_opts,
		png_options:       png_opts,
		SeedSVG:           true,
		SeedPNG:           false,
	}

	return &w, nil
}

func (w *LocalWorker) RenderRasterzenTile(t slippy.Tile) error {

	fh, err := tile.RenderRasterzenTile(t, w.cache, w.nextzen_options, w.rasterzen_options)

	if err != nil {
		return err
	}

	return fh.Close()
}

func (w *LocalWorker) RenderGeoJSONTile(t slippy.Tile) error {

	fh, err := tile.RenderGeoJSONTile(t, w.cache, w.nextzen_options, w.rasterzen_options)

	if err != nil {
		return err
	}

	return fh.Close()
}

func (w *LocalWorker) RenderExtentTile(t slippy.Tile) error {

	fh, err := tile.RenderExtentTile(t, w.cache, w.nextzen_options)

	if err != nil {
		return err
	}

	return fh.Close()
}

func (w *LocalWorker) RenderSVGTile(t slippy.Tile) error {

	fh, err := tile.RenderSVGTile(t, w.cache, w.nextzen_options, w.rasterzen_options, w.svg_options)

	if err != nil {
		return err
	}

	return fh.Close()
}

func (w *LocalWorker) RenderPNGTile(t slippy.Tile) error {

	fh, err := tile.RenderPNGTile(t, w.cache, w.nextzen_options, w.rasterzen_options, w.svg_options, w.png_options)

	if err != nil {
		return err
	}

	return fh.Close()
}
