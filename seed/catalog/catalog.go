package catalog

import (
	"github.com/go-spatial/geom/slippy"
)

type SeedCatalog interface {
	Load(string) (interface{}, bool)
	LoadOrStore(string, slippy.Tile) (interface{}, bool)
	Delete(string)
	Range(func(key, value interface{}) bool) error
	Count() int32
}
