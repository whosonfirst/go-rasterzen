package nextzen

import (
	"crypto/tls"
	"net/http"
)

// this is here to disable HTTP2 in an effort to track down why nextzen
// sometimes returns '400 Bad Request' errors (20181026/thisisaaronland)

func NewHTTPClient() (*http.Client, error) {

	tr := &http.Transport{
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
	}

	client := &http.Client{
		Transport: tr,
	}

	return client, nil
}
