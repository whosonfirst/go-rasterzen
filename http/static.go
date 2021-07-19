package http

import (
	gohttp "net/http"
	"github.com/whosonfirst/go-rasterzen/static"	
)

func StaticFileSystem() (gohttp.FileSystem, error) {
	fs := gohttp.FS(static.FS)
	return fs, nil
}

func StaticHandler() (gohttp.Handler, error) {

	fs, err := StaticFileSystem()

	if err != nil {
		return nil, err
	}
		
	return gohttp.FileServer(fs), nil
}
