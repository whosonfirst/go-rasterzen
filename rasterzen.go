package rasterzen

import (
       "github.com/go-rasterzen/nextzen"
       "github.com/go-rasterzen/tile"       
       "fmt"
       "io"
)

// Something something something GetTileWithLambda... but I'm not really sure what that means
// yet... (20180801/thisisaaronland)
//
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/lambda-go-example-run-function.html
// https://github.com/awsdocs/aws-doc-sdk-examples/blob/master/go/example_code/lambda/aws-go-sdk-lambda-example-run-function.go

GetPNGTileWithCache(c Cache, z int, x int, y int) (io.ReadCloser, error) {

	fh, err := GetSVGTileWithCache(c, z, x, y)

	if err != nil {
	   return nil, err
	}

	wr := "FIX ME"
	
	err := tile.ToPNG(fh, wr)

	if err != nil {
		return nil, err
	}

	return "FIX ME", nil
}

GetSVGTileWithCache(c Cache, z int, x int, y int) (io.ReadCloser, error) {

}

GetRasterzenTileWithCache(c Cache, z int, x int, y int) (io.ReadCloser, error) {

	key := fmt.Sprintf("%d/%d/%d.json", z, x, y)

	nextzen_key := filepath.Join("nextzen", key)
	rasterzen_key := filepath.Join("rasterzen", key)

	var nextzen_data io.ReadCloser   // stuff sent back from nextzen.org
	var rasterzen_data io.ReadCloser // nextzen.org data cropped and manipulated

	rasterzen_data, err = c.Get(rasterzen_key)

	if err == nil {
		return rasterzen_data, nil
	}

	nextzen_data, err = h.Cache.Get(nextzen_key)

	if err != nil {

		// FIX ME... api_key
		
		api_key := "FIX ME"
		
		t, err := nextzen.FetchTile(z, x, y, api_key)

		if err != nil {
			return nil, err
		}

		defer t.Close()

		nextzen_data, err = c.Set(nextzen_key, t)

		if err != nil {
			return nil, err
		}
	}

	cr, err := nextzen.CropTile(z, x, y, nextzen_data)

	if err != nil {
		return nil, err
	}

	defer cr.Close()

	fh, err := c.Set(rasterzen_key, cr)

	return fh, err	  
}