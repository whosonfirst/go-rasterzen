vendor-deps: 
	go mod vendor

fmt:
	go fmt cloudwatch/*.go
	go fmt cmd/*.go
	go fmt config/*.go
	go fmt ecs/*.go
	go fmt lambda/*.go
	go fmt s3/*.go
	go fmt sqs/*.go
	go fmt session/*.go
	go fmt util/*.go

tools:
	@GOPATH=$(GOPATH) go build -o bin/s3 cmd/s3/main.go
	@GOPATH=$(GOPATH) go build -o bin/secret cmd/secret/main.go
	@GOPATH=$(GOPATH) go build -o bin/ecs-run-task cmd/ecs-run-task/main.go
	@GOPATH=$(GOPATH) go build -o bin/lambda-run-task cmd/lambda-run-task/main.go
