CWD=$(shell pwd)

tools:
	rm -rf bin/*
	go build -mod vendor -o bin/rasterd cmd/rasterd/main.go
	go build -mod vendor -o bin/rasterzen-seed cmd/rasterzen-seed/main.go
	go build -mod vendor -o bin/rasterpng cmd/rasterpng/main.go
	go build -mod vendor -o bin/rastersvg cmd/rastersvg/main.go

lambda:
	@make lambda-seed
	@make lambda-seed-sqs

lambda-rasterd:	
	if test -f main; then rm -f main; fi
	if test -f rasterd.zip; then rm -f rasterd.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/rasterd/main.go
	zip rasterd.zip main
	rm -f main

lambda-seed:	
	if test -f main; then rm -f main; fi
	if test -f rasterzen-seed.zip; then rm -f rasterzen-seed.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/rasterd/main.go
	zip rasterzen-seed.zip main
	rm -f main

lambda-seed-sqs:	
	if test -f main; then rm -f main; fi
	if test -f rasterzen-seed-sqs.zip; then rm -f rasterzen-seed-sqs.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/rasterzen-seed-sqs/main.go
	zip rasterzen-seed-sqs.zip main
	rm -f main

docker:
	go mod vendor
	docker build -t rasterzen-seed .
