package main

// see also: docs/rasterzen-seed-sqs-arch.jpg

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-spatial/geom/slippy"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"log"
	"time"
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

			// to do: get NextzenAPI key, etc. from SQSMessage?
			// wrkr, err := worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, s3_cache, nz_opts, svg_opts)

			t := slippy.Tile{Z: msg.Z, X: msg.X, Y: msg.Y}

			t1 := time.Now()

			defer func() {
				log.Printf("time to render %v (%s): %v\n", t, msg.Prefix, time.Since(t1))
			}()

			var render_err error
			
			switch msg.Prefix {
			case "rasterzen":
				render_err = wrkr.RenderRasterzenTile(t)
			case "geojson":
				render_err = wrkr.RenderGeoJSONTile(t)
			case "extent":
				render_err = wrkr.RenderExtentTile(t)
			case "svg":
				render_err = wrkr.RenderSVGTile(t)
			case "png":
				render_err = wrkr.RenderPNGTile(t)
			default:
				render_err = errors.New("Invalid prefix")
			}

			if render_err != nil {
				log.Printf("ERROR failed to render %v (%s): %v\n", t, msg.Prefix, render_err)
			}
		}

		return nil
	}

	return fn, nil
}

func main() {

	nextzen_apikey := flag.String("nextzen-apikey", "", "A valid Nextzen API key.")
	nextzen_origin := flag.String("nextzen-origin", "", "An optional HTTP 'Origin' host to pass along with your Nextzen requests.")
	nextzen_debug := flag.Bool("nextzen-debug", false, "Log requests (to STDOUT) to Nextzen tile servers.")
	nextzen_uri := flag.String("nextzen-uri", "", "A valid URI template (RFC 6570) pointing to a custom Nextzen endpoint.")

	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	custom_svg_options := flag.String("svg-options", "", "The path to a valid RasterzenSVGOptions JSON file.")

	var lambda_dsn flags.DSNString
	flag.Var(&lambda_dsn, "lambda-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'")

	lambda_function := flag.String("lambda-function", "Rasterzen", "A valid AWS Lambda function name.")

	flag.Parse()

	err := flags.SetFlagsFromEnvVars("RASTERZEN_SEED")

	if err != nil {
		log.Fatal(err)
	}

	nz_opts := &nextzen.Options{
		ApiKey: *nextzen_apikey,
		Origin: *nextzen_origin,
		Debug:  *nextzen_debug,
	}

	if *nextzen_uri != "" {

		template, err := uritemplates.Parse(*nextzen_uri)

		if err != nil {
			log.Fatal(err)
		}

		nz_opts.URITemplate = template
	}

	opts, err := s3.NewS3CacheOptionsFromString(*s3_opts)

	if err != nil {
		log.Fatal(err)
	}

	s3_cache, err := s3.NewS3Cache(*s3_dsn, opts)

	if err != nil {
		log.Fatal(err)
	}

	var svg_opts *tile.RasterzenSVGOptions

	if *custom_svg_options != "" {

		opts, err := tile.RasterzenSVGOptionsFromFile(*custom_svg_options)

		if err != nil {
			log.Fatal(err)
		}

		svg_opts = opts

	} else {

		opts, err := tile.DefaultRasterzenSVGOptions()

		if err != nil {
			log.Fatal(err)
		}

		svg_opts = opts
	}

	wrkr, err := worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, s3_cache, nz_opts, svg_opts)

	if err != nil {
		log.Fatal(err)
	}

	handler, err := SQSHandler(wrkr)

	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(handler)
}
