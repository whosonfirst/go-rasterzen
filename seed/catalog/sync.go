package catalog

import (
	"github.com/go-spatial/geom/slippy"
	"sync"
)

type InMemorySeedCatalog struct {
	SeedCatalog
	seed_catalog *sync.Map
}

func NewInMemorySeedCatalog() (SeedCatalog, error) {

	seed_catalog := new(sync.Map)

	m := InMemorySeedCatalog{
		seed_catalog: seed_catalog,
	}

	return &m, nil
}

func (m *InMemorySeedCatalog) LoadOrStore(k string, t slippy.Tile) error {
	m.seed_catalog.LoadOrStore(k, t)
	return nil
}

func (m *InMemorySeedCatalog) Remove(k string) error {
	m.seed_catalog.Delete(k)
	return nil
}

func (m *InMemorySeedCatalog) Range(f func(key, value interface{}) bool) {
	m.seed_catalog.Range(f)
}
