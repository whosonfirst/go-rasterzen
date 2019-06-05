package tile

import (
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"image"
	"io"
	"io/ioutil"
	"os"
)

func RasterzenToImage(in io.Reader) (image.Image, error) {

	tmpfile, err := ioutil.TempFile("", "svg")

	if err != nil {
		return nil, err
	}

	defer func() {

		_, err := os.Stat(tmpfile.Name())

		if !os.IsNotExist(err) {
			os.Remove(tmpfile.Name())
		}
	}()

	err = RasterzenToSVG(in, tmpfile)

	if err != nil {
		return nil, err
	}

	tmpfile.Close()

	return RasterzenPathToImage(tmpfile.Name())
}

func RasterzenPathToImage(path string) (image.Image, error) {

	icon, err := oksvg.ReadIcon(path, oksvg.StrictErrorMode)

	if err != nil {
		return nil, err
	}

	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)

	icon.Draw(raster, 1.0)

	return img, nil
}
