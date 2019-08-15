package catalog

import (
	"github.com/go-spatial/geom/slippy"
)

type SeedCatalog interface {
	LoadOrStore(string, slippy.Tile) error
	Remove(string) error
	Range(func(key, value interface{}) bool)
}
