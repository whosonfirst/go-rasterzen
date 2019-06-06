CWD=$(shell pwd)

vendor-deps:
	go mod vendor

fmt:
	go fmt *.go
	go fmt http/*.go
	go fmt seed/*.go
	go fmt tile/*.go
	go fmt server/*.go
	go fmt nextzen/*.go
	go fmt worker/*.go

# See this? See the release numbers? I don't know why I need to do this
# but apparently 'go mod vendor' excludes things in a package's cmd folder
# Why... I have no idea. Perhaps I am just going it wrong...
# (20190606/thisisaaronland)

bindata:
	if test ! -d cmd/go-bindata; then mkdir -p cmd/go-bindata; fi
	if test ! -d cmd/go-bindata-assetfs; then mkdir -p cmd/go-bindata-assetfs; fi
	curl -s -o cmd/go-bindata/main.go https://raw.githubusercontent.com/whosonfirst/go-bindata/v0.1.0/cmd/go-bindata/main.go
	curl -s -o cmd/go-bindata-assetfs/main.go https://raw.githubusercontent.com/whosonfirst/go-bindata-assetfs/v1.0.1/cmd/go-bindata-assetfs/main.go
	@echo "This file was cloned from https://raw.githubusercontent.com/whosonfirst/go-bindata/v0.1.0/cmd/go-bindata/main.go" > cmd/go-bindata/README.md
	@echo "This file was cloned from https://raw.githubusercontent.com/whosonfirst/go-bindata-assetfs/v0.1.0/cmd/go-bindata-assetfs/main.go" > cmd/go-bindata-assetfs/README.md

html:
	go build -mod vendor -o bin/go-bindata cmd/go-bindata/main.go
	go build -mod vendor -o bin/go-bindata-assetfs cmd/go-bindata-assetfs/main.go
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
	@make html
	@make tools

lambda:	
	if test -f main; then rm -f main; fi
	if test -f deployment.zip; then rm -f deployment.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/rasterd.go
	zip deployment.zip main
	rm -f main
