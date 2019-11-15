package worker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "log"
	"strings"
)

type seedResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type seedRequest struct {
	Path        string           `json:"path"`
	HTTPMethod  string           `json:"httpMethod"`
	QueryString seedRequestQuery `json:"queryStringParameters"`
}

type seedRequestQuery struct {
	ApiKey  string `json:"api_key"`
	Discard string `json:"discard,omitempty"`
}

type LambdaWorker struct {
	Worker
	function          string
	client            *lambda.Lambda
	cache             cache.Cache
	nextzen_options   *nextzen.Options
	rasterzen_options *tile.RasterzenOptions
	svg_options       *tile.RasterzenSVGOptions
	png_options       *tile.RasterzenPNGOptions
}

func NewLambdaWorker(dsn map[string]string, function string, c cache.Cache, nz_opts *nextzen.Options, rz_opts *tile.RasterzenOptions, svg_opts *tile.RasterzenSVGOptions, png_opts *tile.RasterzenPNGOptions) (Worker, error) {

	creds, ok := dsn["credentials"]

	if !ok {
		return nil, errors.New("Missing credentials")
	}

	region, ok := dsn["region"]

	if !ok {
		return nil, errors.New("Missing region")
	}

	sess, err := session.NewSessionWithCredentials(creds, region)

	if err != nil {
		return nil, err
	}

	client := lambda.New(sess)

	w := LambdaWorker{
		client:            client,
		function:          function,
		cache:             c,
		nextzen_options:   nz_opts,
		rasterzen_options: rz_opts,
		svg_options:       svg_opts,
		png_options:       png_opts,
	}

	return &w, nil
}

func (w *LambdaWorker) RenderRasterzenTile(t slippy.Tile) error {
	return w.renderTile(t, "rasterzen", "json")
}

func (w *LambdaWorker) RenderGeoJSONTile(t slippy.Tile) error {
	return w.renderTile(t, "geojson", "geojson")
}

func (w *LambdaWorker) RenderExtentTile(t slippy.Tile) error {
	return w.renderTile(t, "extent", "geojson")
}

func (w *LambdaWorker) RenderSVGTile(t slippy.Tile) error {
	return w.renderTile(t, "svg", "svg")
}

func (w *LambdaWorker) RenderPNGTile(t slippy.Tile) error {
	return w.renderTile(t, "png", "png")
}

func (w *LambdaWorker) renderTile(t slippy.Tile, prefix, format string) error {

	// handle refresh stuff here...

	cache_key := tile.CacheKeyForTile(t, prefix, format)

	cached, err := CheckCache(w.cache, cache_key)

	if err != nil {
		return err
	}

	if cached {
		return nil
	}

	uri := fmt.Sprintf("/%s", cache_key)

	api_key := w.nextzen_options.ApiKey

	query := seedRequestQuery{
		ApiKey: api_key,
	}

	// if we are a (local) null cache then don't bother asking the
	// lambda/proxy endpoint to return any data across the wire
	// (20181105/thisisaaronland)

	if w.cache.Name() == "multi#null" {
		query.Discard = "1"
	}

	req := seedRequest{
		HTTPMethod:  "GET",
		Path:        uri,
		QueryString: query,
	}

	payload, err := json.Marshal(req)

	if err != nil {
		return err
	}

	input := &lambda.InvokeInput{
		FunctionName: aws.String(w.function),
		Payload:      payload,
	}

	aws_rsp, err := w.client.Invoke(input)

	if err != nil {
		return err
	}

	// https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#InvokeOutput
	// {"statusCode":200,"headers":{"Access-Control-Allow-Origin":"*","Content-Type":"image/svg+xml"},"body": ...

	if *aws_rsp.StatusCode != int64(200) {
		msg := fmt.Sprintf("Lambda invocation error: %d (%v)", aws_rsp.StatusCode, aws_rsp.FunctionError)
		return errors.New(msg)
	}

	if aws_rsp.FunctionError != nil {
		msg := fmt.Sprintf("Lambda function error: %s", *aws_rsp.FunctionError)
		return errors.New(msg)
	}

	var rsp seedResponse

	err = json.Unmarshal(aws_rsp.Payload, &rsp)

	if err != nil {
		return err
	}

	var r io.Reader

	if format == "png" {

		data, err := base64.StdEncoding.DecodeString(rsp.Body)

		if err != nil {
			return err
		}

		r = bytes.NewReader(data)

	} else {

		r = strings.NewReader(rsp.Body)
	}

	fh := ioutil.NopCloser(r)

	cache_fh, err := w.cache.Set(cache_key, fh)

	if err != nil {
		return nil
	}

	return cache_fh.Close()
}
