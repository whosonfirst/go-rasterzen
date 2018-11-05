package seed

// WORK IN PROGRESS
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/lambda-go-example-run-function.html

import (
	"encoding/json"	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/go-spatial/geom/slippy"	
	"log"
)

type SeedRequest struct {
}

func SeedTileLambda(s *TileSeeder, t slippy.Tile) error {

	region := "us-west-2"
	function := "Rasterzen"

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := lambda.New(sess, &aws.Config{Region: aws.String(region)})

	request := SeedRequest{}

	payload, err := json.Marshal(request)

	if err != nil {
		return err
	}

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
