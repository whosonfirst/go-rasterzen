package catalog

import (
	"github.com/go-spatial/geom/slippy"
	"sync"
)

type SyncMapSeedCatalog struct {
	SeedCatalog
	seed_catalog *sync.Map
}

func NewSyncMapSeedCatalog() (SeedCatalog, error) {

	seed_catalog := new(sync.Map)

	m := SyncMapSeedCatalog{
		seed_catalog: seed_catalog,
	}

	return &m, nil
}

func (m *SyncMapSeedCatalog) LoadOrStore(k string, t slippy.Tile) error {
	m.seed_catalog.LoadOrStore(k, t)
	return nil
}

func (m *SyncMapSeedCatalog) Remove(k string) error {
	m.seed_catalog.Delete(k)
	return nil
}

func (m *SyncMapSeedCatalog) Range(f func(key, value interface{}) bool) error {
	m.seed_catalog.Range(f)
	return nil
}
