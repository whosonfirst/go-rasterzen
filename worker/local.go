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

func (w *LocalWorker) SeedTile(t slippy.Tile) error {

	if !w.SeedSVG && !w.SeedPNG {

		_, err := tile.SeedRasterzen(t, w.cache, w.nextzen_options)

		if err != nil {
			return err
		}
	}

	if w.SeedSVG {

		_, err := tile.SeedSVG(t, w.cache, w.nextzen_options)

		if err != nil {
			return err
		}

	}

	if w.SeedPNG {

		_, err := tile.SeedPNG(t, w.cache, w.nextzen_options)

		if err != nil {
			return err
		}

	}

	return nil
}
