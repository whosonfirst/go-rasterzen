package server

import (
	"github.com/whosonfirst/algnhsa"
	_ "log"
	"net/http"
	 "net/url"
)

type LambdaServer struct {
	Server
	url *url.URL
}

func NewLambdaServer(u *url.URL, args ...interface{}) (Server, error) {

	server := LambdaServer{
		url: u,
	}

	return &server, nil
}

func (s *LambdaServer) Address() string {
	return s.url.String()
}

func (s *LambdaServer) ListenAndServe(mux *http.ServeMux) error {

        // this cr^H^H^H stuff is important (20180713/thisisaaronland)
	// go-whosonfirst-static/README.md#lambda-api-gateway-and-images#lambda-api-gateway-and-images
	
	lambda_opts := new(algnhsa.Options)
	lambda_opts.BinaryContentTypes = []string{"image/png"}

	algnhsa.ListenAndServe(mux, lambda_opts)
	return nil
}
