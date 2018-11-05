package seed

// WORK IN PROGRESS
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/lambda-go-example-run-function.html

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/go-spatial/geom/slippy"
	"log"
	"net/url"
)

type SeedRequest struct {
	Resource    string      `json:"resource"`
	Path        string      `json:"path"`
	HTTPMethod  string      `json:"httpMethod"`
	QueryString interface{} `json:"queryStringParameters"`
}

func NewSeedRequest(s *TileSeeder, t slippy.Tile) (*SeedRequest, error) {

	api_key := s.NextzenOptions.ApiKey

	format := "svg"
	uri := fmt.Sprintf("/%s/%d/%d/%d.%s", format, t.Z, t.X, t.Y, format)

	q := url.Values{}
	q.Add("api_key", api_key)

	req := SeedRequest{
		Resource:    "/{proxy+}",
		HTTPMethod:  "GET",
		Path:        uri,
		QueryString: uri,
	}

	return &req, nil
}

func SeedTileLambda(s *TileSeeder, t slippy.Tile) error {

	creds := "session"
	region := "us-west-2"
	
	function := "RasterzenDebug"
	
	sess, err := session.NewSessionWithCredentials(creds, region)

	/*
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := lambda.New(sess, &aws.Config{Region: aws.String(region)})
	*/

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

	rsp, err := client.Invoke(input)

	if err != nil {
		return err
	}

	log.Println(rsp)
	return nil
}
