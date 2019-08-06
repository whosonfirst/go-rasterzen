package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	_ "fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"log"
)

type SQSHandlerFunc func(ctx context.Context, sqsEvent events.SQSEvent) error

func SQSHandler(wrkr worker.Worker) (SQSHandlerFunc, error) {

	fn := func(ctx context.Context, sqsEvent events.SQSEvent) error {

		for _, message := range sqsEvent.Records {

			body := message.Body
			var msg worker.SQSMessage

			err := json.Unmarshal([]byte(body), &msg)

			if err != nil {
				return err
			}

			t := slippy.Tile{Z: msg.Z, X: msg.X, Y: msg.Y}

			switch msg.Prefix {
			case "rasterzen":
				wrkr.RenderRasterzenTile(t)
			case "geojson":
				wrkr.RenderGeoJSONTile(t)
			case "extent":
				wrkr.RenderExtentTile(t)
			case "svg":
				wrkr.RenderSVGTile(t)
			case "png":
				wrkr.RenderPNGTile(t)
			default:
				return errors.New("Invalid prefix")
			}

		}

		return nil
	}

	return fn, nil
}

func main() {

	var lambda_dsn flags.DSNString
	flag.Var(&lambda_dsn, "lambda-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'")

	lambda_function := flag.String("lambda-function", "Rasterzen", "A valid AWS Lambda function name.")

	flag.Parse()

	wrkr, err := worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, c, nz_opts, svg_opts)

	if err != nil {
		log.Fatal(err)
	}

	handler, err := SQSHandler(wrkr)

	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(handler)
}
