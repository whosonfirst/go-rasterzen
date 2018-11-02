package http

import (
	gohttp "net/http"
)

func StaticFileSystem() (gohttp.FileSystem, error) {
	fs := assetFS()
	return fs, nil
}

func StaticHandler() (gohttp.Handler, error) {

	fs := assetFS()
	return gohttp.FileServer(fs), nil
}
