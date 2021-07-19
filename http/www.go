package http

import (
	"github.com/whosonfirst/go-rasterzen/templates/html"
	gohttp "net/http"
	"html/template"
)

type WWWHandlerOptions struct {
	NextzenAPIKey string
	Debug         bool
	TileFormat    string
}

type WWWTemplateVars struct {
	NextzenAPIKey string
	Debug         bool
	TileFormat    string
}

func WWWHandler(opts *WWWHandlerOptions) (gohttp.HandlerFunc, error) {

	t := template.New("index")
	t, err := t.ParseFS(html.FS, "*.html")

	if err != nil {
		return nil, err
	}

	vars := WWWTemplateVars{
		NextzenAPIKey: opts.NextzenAPIKey,
		Debug:         opts.Debug,
		TileFormat:    opts.TileFormat,
	}

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		err = t.Execute(rsp, vars)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	return gohttp.HandlerFunc(fn), nil
}
