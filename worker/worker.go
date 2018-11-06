package worker

import (
	"github.com/go-spatial/geom/slippy"
)

type Worker interface {
	RenderRasterzenTile(slippy.Tile) error
	RenderSVGTile(slippy.Tile) error
	RenderPNGTile(slippy.Tile) error
}
