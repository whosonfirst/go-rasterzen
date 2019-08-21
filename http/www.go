package http

import (
	"github.com/whosonfirst/go-bindata-html-template"
	"github.com/whosonfirst/go-rasterzen/assets/templates"
	gohttp "net/http"
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

	t := template.New("index", templates.Asset)
	// t = t.Funcs(funcs)

	t, err := t.ParseFiles("templates/html/index.html")

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
