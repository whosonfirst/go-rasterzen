package tile

import (
	"fmt"
	"github.com/go-spatial/geom/slippy"
)

func CacheKeyForRasterzenTile(t slippy.Tile) string {
	return CacheKeyForTile(t, "rasterzen", "json")
}

func CacheKeyForPNGTile(t slippy.Tile) string {
	return CacheKeyForTile(t, "png", "png")
}

func CacheKeyForSVGTile(t slippy.Tile) string {
	return CacheKeyForTile(t, "svg", "svg")
}

func CacheKeyForTile(t slippy.Tile, prefix string, ext string) string {

	return fmt.Sprintf("%s/%d/%d/%d.%s", prefix, t.Z, t.X, t.Y, ext)
}
