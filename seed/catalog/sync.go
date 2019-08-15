package catalog

import (
	"github.com/go-spatial/geom/slippy"
	"sync"
	"sync/atomic"
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

func (m *SyncMapSeedCatalog) Count() int32 {

	remaining := int32(0)

	tile_func := func(key, value interface{}) bool {
		atomic.AddInt32(&remaining, 1)
		return true
	}

	m.Range(tile_func)

	return remaining
}
