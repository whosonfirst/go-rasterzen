vendor-deps:
	go mod vendor

fmt:
	go fmt ./...

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
	go build -o bin/rasterd cmd/rasterd/main.go
	go build -o bin/rasterzen-seed cmd/rasterzen-seed/main.go
	go build -o bin/rasterpng cmd/rasterpng/main.go

rebuild:
	@make assets
	@make tools

lambda:	
	if test -f main; then rm -f main; fi
	if test -f deployment.zip; then rm -f deployment.zip; fi
	GOOS=linux go build -o main cmd/rasterd.go
	zip deployment.zip main
	rm -f main
