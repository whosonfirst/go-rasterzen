CWD=$(shell pwd)

bake-assets:
	go build -o bin/go-bindata cmd/go-bindata/main.go
	go build -o bin/go-bindata-assetfs cmd/go-bindata-assetfs/main.go
	rm -f www/static/*~ www/static/css/*~ www/static/javascript/*~
	@PATH=$(PATH):$(CWD)/bin bin/go-bindata-assetfs -pkg http static/javascript static/css
	mv bindata.go http/assetfs.go
	rm -rf templates/html/*~
	bin/go-bindata -pkg templates -o assets/templates/html.go templates/html

tools:
	rm -rf bin/*
	go build -mod vendor -o bin/rasterd cmd/rasterd/main.go
	go build -mod vendor -o bin/rasterzen-seed cmd/rasterzen-seed/main.go
	go build -mod vendor -o bin/rasterpng cmd/rasterpng/main.go
	go build -mod vendor -o bin/rastersvg cmd/rastersvg/main.go

rebuild:
	@make assets
	@make tools

lambda:
	@make lambda-seed
	@make lambda-seed-sqs

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
