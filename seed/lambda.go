package seed

// WORK IN PROGRESS
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/lambda-go-example-run-function.html

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"log"
	"os"
	"strconv"
)

type seedRequest struct {
	Resource    string      `json:"resource"`
	Path        string      `json:"path"`
	HTTPMethod  string      `json:"httpMethod"`
	QueryString seedRequestQuery `json:"queryStringParameters"`	
}

type seedRequestQuery struct {
	ApiKey string `json:"api_key"`
}

type seedResponseError struct {
	Message string `json:"message"`
}

type seedResponseData struct {
	Item string `json:"item"`
}

type seedResponseBody struct {
	Result string             `json:"result"`
	Data   []seedResponseData `json:"data"`
	Error  seedResponseError  `json:"error"`
}

type seedResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type seedResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    seedResponseHeaders `json:"headers"`
	Body       seedResponseBody    `json:"body"`
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

	log.Println("RSP", aws_rsp, err)

	var rsp seedResponse

	err = json.Unmarshal(aws_rsp.Payload, &rsp)

	if err != nil {
		return err
	}

	// If the status code is NOT 200, the call failed
	if rsp.StatusCode != 200 {
		fmt.Println("Error getting items, StatusCode: " + strconv.Itoa(rsp.StatusCode))
		os.Exit(0)
	}

	// If the result is failure, we got an error
	if rsp.Body.Result == "failure" {
		fmt.Println("Failed to get items")
		os.Exit(0)
	}

	// Print out items
	if len(rsp.Body.Data) > 0 {
		for i := range rsp.Body.Data {
			fmt.Println(rsp.Body.Data[i].Item)
		}
	} else {
		fmt.Println("There were no items")
	}

	return nil
}
