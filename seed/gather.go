package seed

import (
	"context"
	"errors"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	_ "log"
	"strconv"
	"strings"
	"time"
)

type GatherTilesFunc func(context.Context, *TileSet) error

func NewGatherTilesFunc(tiles []string) (GatherTilesFunc, error) {

	gather_func := func(ctx context.Context, tileset *TileSet) error {

		for _, str_zxy := range tiles {

			z, x, y, err := parse_zxy(str_zxy)

			if err != nil {
				return err
			}

			t := slippy.Tile{
				Z: uint(z),
				X: uint(x),
				Y: uint(y),
			}

			tileset.AddTile(t)
		}

		return nil
	}

	return gather_func, nil
}

func NewGatherTilesExtentFunc(extents []string, sep string, min_zoom int, max_zoom int) (GatherTilesFunc, error) {

	gather_func := func(ctx context.Context, tileset *TileSet) error {

		for _, str_extent := range extents {

			ex, err := parse_extent(str_extent, sep)

			if err != nil {
				return err
			}

			for z := min_zoom; z <= max_zoom; z++ {

				for _, t := range slippy.FromBounds(ex, uint(z)) {
					tileset.AddTile(t)
				}
			}
		}

		return nil
	}

	return gather_func, nil
}

func GatherTiles(tileset *TileSet, seeder *TileSeeder, f GatherTilesFunc) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if tileset.Timings {

		t1 := time.Now()

		ticker := time.NewTicker(time.Second * 5)

		defer func() {
			tileset.Logger.Status("Time to gather %d tiles: %v\n", tileset.ToSeed, time.Since(t1))
		}()

		go func() {

			for range ticker.C {

				select {
				case <-ctx.Done():
					return
				default:
					tileset.Logger.Status("Still gathering tiles, %d so far (%v)", tileset.ToSeed, time.Since(t1))
				}
			}
		}()

	}

	return f(ctx, tileset)
}

func parse_extent(str_extent string, sep string) (*geom.Extent, error) {

	coords := strings.Split(str_extent, sep)

	if len(coords) != 4 {
		return nil, errors.New("Invalid string extent")
	}

	minx, err := strconv.ParseFloat(coords[0], 64)

	if err != nil {
		return nil, err
	}

	miny, err := strconv.ParseFloat(coords[1], 64)

	if err != nil {
		return nil, err
	}

	maxy, err := strconv.ParseFloat(coords[3], 64)

	if err != nil {
		return nil, err
	}

	maxx, err := strconv.ParseFloat(coords[2], 64)

	if err != nil {
		return nil, err
	}

	min := [2]float64{minx, miny}
	max := [2]float64{maxx, maxy}

	return geom.NewExtent(min, max), nil
}

func parse_zxy(str_zxy string) (int, int, int, error) {

	parts := strings.Split(str_zxy, "/")

	if len(parts) != 3 {
		return 0, 0, 0, errors.New("Invalid xzy string")
	}

	z, err := strconv.Atoi(parts[0])

	if err != nil {
		return 0, 0, 0, err
	}

	x, err := strconv.Atoi(parts[1])

	if err != nil {
		return 0, 0, 0, err
	}

	y, err := strconv.Atoi(parts[2])

	if err != nil {
		return 0, 0, 0, err
	}

	return z, x, y, nil
}
