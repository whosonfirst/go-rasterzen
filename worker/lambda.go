package worker

import (
	"bytes"
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
	"io/ioutil"
	_ "log"
)

type seedResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

type seedRequest struct {
	Resource    string           `json:"resource"`
	Path        string           `json:"path"`
	HTTPMethod  string           `json:"httpMethod"`
	QueryString seedRequestQuery `json:"queryStringParameters"`
}

type seedRequestQuery struct {
	ApiKey string `json:"api_key"`
}

type LambdaWorker struct {
	Worker
	function        string
	client          *lambda.Lambda
	cache           cache.Cache
	nextzen_options *nextzen.Options
}

func NewLambdaWorker(dsn map[string]string, function string, c cache.Cache, nz_opts *nextzen.Options) (Worker, error) {

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
		client:   client,
		function: function,
		cache:    c,
	}

	return &w, nil
}

func (w *LambdaWorker) SeedTile(t slippy.Tile) error {

	return w.seedTile(t, "svg")
}

func (w *LambdaWorker) seedTile(t slippy.Tile, format string) error {

	api_key := w.nextzen_options.ApiKey

	key := tile.CacheKey(t, format, format)
	uri := fmt.Sprintf("/%s", key)

	query := seedRequestQuery{
		ApiKey: api_key,
	}

	req := seedRequest{
		Resource:    "/{proxy+}",
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
		msg := fmt.Sprintf("Lambda err %d (%s)", aws_rsp.StatusCode, aws_rsp.FunctionError)
		return errors.New(msg)
	}

	var rsp seedResponse

	err = json.Unmarshal(aws_rsp.Payload, &rsp)

	if err != nil {
		return err
	}

	r := bytes.NewReader(rsp.Body)
	fh := ioutil.NopCloser(r)

	_, err = w.cache.Set(key, fh)
	return err
}
