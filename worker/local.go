package worker

import (
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
)

type LocalWorker struct {
	Worker
	NextzenOptions *nextzen.Options
}

func NewLocalWorker(opts *SeedOptions) (Worker, error) {

	w := LocalWorkers{
		options: opts,
	}

	return &w, nil
}

func (w *LocalWorker) SeedTile(t slippy.Tile, c cache.Cache) error {

	if !w.opts.SeedSVG && !w.opts.SeedPNG {

		_, err := tile.SeedRasterzen(t, c, w.NextzenOptions)

		if err != nil {
			return err
		}
	}

	if w.SeedSVG {

		_, err := tile.SeedSVG(t, c, w.NextzenOptions)

		if err != nil {
			return err
		}

	}

	if w.SeedPNG {

		_, err := tile.SeedPNG(t, c, w.NextzenOptions)

		if err != nil {
			return err
		}

	}

	return nil
}
