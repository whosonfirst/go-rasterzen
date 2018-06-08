package nextzen

import (
       "fmt"
       "io"
       _ "log"
       "net/http"
)

// THIS SIGNATURE WILL CHANGE - YES
// ALSO PLEASE CACHE ME...

func FetchTile(z int, x int, y int, api_key string) (io.ReadCloser, error) {

     url := fmt.Sprintf("https://tile.nextzen.org/tilezen/vector/v1/256/all/%d/%d/%d.json?api_key=%s", z, x, y, api_key)

     rsp, err := http.Get(url)

     if err != nil {
     	return nil, err
     }

     return rsp.Body, nil
}
