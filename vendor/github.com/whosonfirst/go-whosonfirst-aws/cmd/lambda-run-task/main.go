package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-aws/lambda"	
	"log"
)

func main() {

	var lambda_dsn = flag.String("lambda-dsn", "", "A valid (go-whosonfirst-aws) Lambda DSN.")
	var lambda_func = flag.String("lambda-func", "", "A valid AWS Lambda function name.")
	var lambda_type = flag.String("lambda-type", "RequestResponse", "A valid (go-aws-sdk) lambda.InvocationType string.")

	flag.Parse()
	
	svc, err := lambda.NewLambdaServiceWithDSN(*lambda_dsn)
	
	if err != nil {
		log.Fatal(err)
	}

	// see what's going on here? this should be updated to support more
	// nuanced payloads than just arg1, arg2, etc. (20190225/thisisaaronland)
	
	for _, payload := range flag.Args() {
		
		out, err := lambda.InvokeFunction(svc, *lambda_func, *lambda_type, payload)
		
		if err != nil {
			log.Fatal(err)
		}

		log.Println(out)
	}
}	
