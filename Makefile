assets:
	go build -o bin/go-bindata ./vendor/github.com/whosonfirst/go-bindata/go-bindata/
	go build -o bin/go-bindata-assetfs vendor/github.com/whosonfirst/go-bindata-assetfs/go-bindata-assetfs/main.go
	rm -f www/static/*~ www/static/css/*~ www/static/javascript/*~
	@PATH=$(PATH):$(CWD)/bin bin/go-bindata-assetfs -pkg http static/javascript static/css
	mv bindata.go http/assetfs.go
	rm -rf templates/html/*~
	bin/go-bindata -pkg templates -o assets/templates/html.go templates/html

tools:
	rm -rf bin/*
	go build -mod vendor -o bin/rasterd cmd/rasterd/main.go
	go build -mod vendor -o bin/rasterzen-seed cmd/rasterzen-seed/main.go
	# go build -mod vendor -o bin/rasterzen-seed-sqs cmd/rasterzen-seed-sqs/main.go
	go build -mod vendor -o bin/rasterpng cmd/rasterpng/main.go
	go build -mod vendor -o bin/rastersvg cmd/rastersvg/main.go

rebuild:
	@make assets
	@make tools

lambda:	
	if test -f main; then rm -f main; fi
	if test -f deployment.zip; then rm -f deployment.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/rasterd/main.go
	zip deployment.zip main
	rm -f main

lambda-sqs:	
	if test -f main; then rm -f main; fi
	if test -f deployment.zip; then rm -f deployment.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/rasterzen-seed-sqs/main.go
	zip deployment.zip main
	rm -f main
