package http

import (
	gohttp "net/http"
	"github.com/whosonfirst/go-bindata-html-template"
	"github.com/whosonfirst/go-rasterzen/assets/templates"
)

func WWWHandler(apikey string) (gohttp.HandlerFunc, error) {

	t := template.New("index", templates.Asset)
	// t = t.Funcs(funcs)

	t, err := t.ParseFiles("templates/html/index.html")

	if err != nil {
		return nil, err
	}

	type TemplateVars struct {
		NextzenAPIKey string
	}

	vars := TemplateVars{
		NextzenAPIKey: apikey,
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
