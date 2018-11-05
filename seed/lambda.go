package seed

// WORK IN PROGRESS
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/lambda-go-example-run-function.html

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"log"
)

type seedRequest struct {
	Resource    string           `json:"resource"`
	Path        string           `json:"path"`
	HTTPMethod  string           `json:"httpMethod"`
	QueryString seedRequestQuery `json:"queryStringParameters"`
}

type seedRequestQuery struct {
	ApiKey string `json:"api_key"`
}

func NewSeedRequest(s *TileSeeder, t slippy.Tile) (*seedRequest, error) {

	api_key := s.NextzenOptions.ApiKey

	format := "svg"
	uri := fmt.Sprintf("/%s/%d/%d/%d.%s", format, t.Z, t.X, t.Y, format)

	query := seedRequestQuery{
		ApiKey: api_key,
	}

	req := seedRequest{
		Resource:    "/{proxy+}",
		HTTPMethod:  "GET",
		Path:        uri,
		QueryString: query,
	}

	return &req, nil
}

func SeedTileLambda(s *TileSeeder, t slippy.Tile) error {

	creds := "session"
	region := "us-west-2"

	function := "RasterzenDebug"

	sess, err := session.NewSessionWithCredentials(creds, region)

	if err != nil {
		return err
	}

	client := lambda.New(sess)

	request, err := NewSeedRequest(s, t)

	if err != nil {
		return err
	}

	payload, err := json.Marshal(request)

	if err != nil {
		return err
	}

	log.Println(string(payload))

	input := &lambda.InvokeInput{
		FunctionName: aws.String(function),
		Payload:      payload,
	}

	aws_rsp, err := client.Invoke(input)

	if err != nil {
		return err
	}

	// https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#InvokeOutput

	if *aws_rsp.StatusCode != int64(200) {
		msg := fmt.Sprintf("Lambda err %d (%s)", aws_rsp.StatusCode, aws_rsp.FunctionError)
		return errors.New(msg)
	}

	// {"statusCode":200,"headers":{"Access-Control-Allow-Origin":"*","Content-Type":"image/svg+xml"},"body": ...
	// log.Println("WHAT", string(aws_rsp.Payload))

	return nil
}
