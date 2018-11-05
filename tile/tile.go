package tile

import (
	"fmt"
	"github.com/go-spatial/geom/slippy"
)

func CacheKeyForTile(t slippy.Tile, prefix string, ext string) string {

	return fmt.Sprintf("%s/%d/%d/%d.%s", prefix, t.Z, t.X, t.Y, ext)
}
