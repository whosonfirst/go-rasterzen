package worker

import (
	"github.com/go-spatial/geom/slippy"
)

type Worker interface {
	SeedTile(slippy.Tile) error
}
