package nextzen

import (
	"fmt"
	_ "github.com/paulmach/orb"
	_ "github.com/paulmach/orb/clip"
	_ "github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"io"
	"log"
	"net/http"
)

// THIS SIGNATURE WILL CHANGE - YES
// ALSO PLEASE CACHE ME...

func FetchTile(z int, x int, y int, api_key string) (io.ReadCloser, error) {

	layer := "all"

	url := fmt.Sprintf("https://tile.nextzen.org/tilezen/vector/v1/256/%s/%d/%d/%d.json?api_key=%s", layer, z, x, y, api_key)

	rsp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	zm := maptile.Zoom(uint32(z))
	tl := maptile.Tile{uint32(z), uint32(x), zm}

	log.Println("BOUNDS", tl.Bound())

	// CLIP TO BOUNDS HERE...
	// fc, _ := geojson.UnmarshalFeature([]byte(str_f))
	// clipped := clip.Geometry(bounds, fc.Geometry)

	return rsp.Body, nil
}
