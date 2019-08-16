package catalog

import (
	"encoding/json"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	_ "log"
	"sync"
)

type TileRecord struct {
	Key  string      `json:"key"`
	Tile slippy.Tile `json:"tile"`
}

type SeedCatalogTable struct {
	sqlite.Table
	name string
}

func NewSeedCatalogTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	t, err := NewSeedCatalogTable()

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSeedCatalogTable() (sqlite.Table, error) {

	t := SeedCatalogTable{
		name: "catalog",
	}

	return &t, nil
}

func (t *SeedCatalogTable) Name() string {
	return t.name
}

func (t *SeedCatalogTable) Schema() string {

	sql := `CREATE TABLE %s (
		key TEXT NOT NULL PRIMARY KEY,
		tile TEXT NOT NULL
	);`

	return fmt.Sprintf(sql, t.Name())
}

func (t *SeedCatalogTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *SeedCatalogTable) IndexRecord(db sqlite.Database, i interface{}) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	tr := i.(TileRecord)

	enc_tile, err := json.Marshal(tr.Tile)

	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		key, tile
	) VALUES (
		?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(tr.Key, string(enc_tile))

	if err != nil {
		return err
	}

	return tx.Commit()
}

type SQLiteSeedCatalog struct {
	SeedCatalog
	db    *database.SQLiteDatabase
	table sqlite.Table
	mu    *sync.RWMutex
}

func NewSQLiteSeedCatalog(dsn string) (SeedCatalog, error) {

	db, err := database.NewDB(dsn)

	if err != nil {
		return nil, err
	}

	tbl, err := NewSeedCatalogTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	mu := new(sync.RWMutex)

	m := SQLiteSeedCatalog{
		db:    db,
		table: tbl,
		mu:    mu,
	}

	return &m, nil
}

func (m *SQLiteSeedCatalog) LoadOrStore(k string, t slippy.Tile) error {

	tile_record := TileRecord{
		Key:  k,
		Tile: t,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.table.IndexRecord(m.db, tile_record)
}

func (m *SQLiteSeedCatalog) Remove(k string) error {

	m.db.Lock()
	defer m.db.Unlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := m.db.Conn()

	if err != nil {
		return err
	}

	_, err = conn.Exec("DELETE FROM catalog WHERE key = ?", k)

	if err != nil {
		return err
	}

	return nil
}

func (m *SQLiteSeedCatalog) Range(f func(key, value interface{}) bool) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := m.db.Conn()

	if err != nil {
		return err
	}

	rows, err := conn.Query(fmt.Sprintf("SELECT * FROM %s", m.table.Name()))

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {

		var key string
		var enc_tile string

		err = rows.Scan(&key, &enc_tile)

		if err != nil {
			return err
		}

		var t slippy.Tile

		err = json.Unmarshal([]byte(enc_tile), &t)

		if err != nil {
			return err
		}

		f(key, t)
	}

	err = rows.Err()

	if err != nil {
		return err
	}

	return nil
}

func (m *SQLiteSeedCatalog) Count() int32 {

	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := m.db.Conn()

	if err != nil {
		return -1
	}

	query := fmt.Sprintf("SELECT COUNT(key) FROM %s", m.table.Name())
	row := conn.QueryRow(query)

	var count int32
	err = row.Scan(&count)

	if err != nil {
		return -1
	}

	return count
}
