CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src; then mkdir src; fi
	mkdir -p src/github.com/whosonfirst/go-rasterzen
	cp *.go src/github.com/whosonfirst/go-rasterzen/
	cp -r assets src/github.com/whosonfirst/go-rasterzen/
	cp -r nextzen src/github.com/whosonfirst/go-rasterzen/
	cp -r tile src/github.com/whosonfirst/go-rasterzen/
	cp -r http src/github.com/whosonfirst/go-rasterzen/
	cp -r seed src/github.com/whosonfirst/go-rasterzen/
	cp -r server src/github.com/whosonfirst/go-rasterzen/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/paulmach/orb"
	@GOPATH=$(GOPATH) go get -u "github.com/srwiley/oksvg"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/sjson"
	@GOPATH=$(GOPATH) go get -u "github.com/go-spatial/geom"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/geojson2svg/pkg/geojson2svg"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-cache-s3"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-cli"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-aws"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/algnhsa"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-bindata"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-bindata-assetfs"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-bindata-html-template"
	@GOPATH=$(GOPATH) go get -u "github.com/jtacoma/uritemplates"
	rm -rf src/github.com/whosonfirst/go-bindata/testdata
	# mv src/github.com/whosonfirst/go-whosonfirst-cache-s3/vendor/github.com/whosonfirst/go-whosonfirst-aws src/github.com/whosonfirst/
	# mv src/github.com/whosonfirst/go-whosonfirst-cache-s3/vendor/github.com/aws/aws-sdk-go src/github.com/aws/
	mv src/github.com/whosonfirst/go-whosonfirst-aws/vendor/github.com/aws/aws-sdk-go src/github.com/aws/
	rm -rf src/github.com/whosonfirst/go-whosonfirst-cache-s3/vendor/github.com/whosonfirst/go-whosonfirst-aws
	rm -rf src/github.com/whosonfirst/go-whosonfirst-cache-s3/vendor/github.com/aws/aws-sdk-go
	mv src/github.com/whosonfirst/go-whosonfirst-cache-s3/vendor/github.com/whosonfirst/go-whosonfirst-cache src/github.com/whosonfirst/

# if you're wondering about the 'rm -rf' stuff below it's because Go is
# weird... https://vanduuren.xyz/2017/golang-vendoring-interface-confusion/
# (20170912/thisisaaronland)

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt cmd/*.go
	go fmt http/*.go
	go fmt seed/*.go
	go fmt tile/*.go
	go fmt server/*.go
	go fmt nextzen/*.go

# See notes in cmd/rasterd.go (20181102/thisisaaronland)

assets:	self
	@GOPATH=$(GOPATH) go build -o bin/go-bindata ./vendor/github.com/whosonfirst/go-bindata/go-bindata/
	@GOPATH=$(GOPATH) go build -o bin/go-bindata-assetfs vendor/github.com/whosonfirst/go-bindata-assetfs/go-bindata-assetfs/main.go
	rm -f www/static/*~ www/static/css/*~ www/static/javascript/*~
	@PATH=$(PATH):$(CWD)/bin bin/go-bindata-assetfs -pkg http static/javascript static/css
	mv bindata.go http/assetfs.go
	rm -rf templates/html/*~
	@GOPATH=$(GOPATH) bin/go-bindata -pkg templates -o assets/templates/html.go templates/html

bin: 	self
	rm -rf bin/*
	@GOPATH=$(GOPATH) go build -o bin/rasterd cmd/rasterd.go
	@GOPATH=$(GOPATH) go build -o bin/rasterzen-seed cmd/rasterzen-seed.go

rebuild:
	@make assets
	@make bin

lambda:	
	@make self
	if test -f main; then rm -f main; fi
	if test -f deployment.zip; then rm -f deployment.zip; fi
	@GOPATH=$(GOPATH) GOOS=linux go build -o main cmd/rasterd.go
	zip deployment.zip main
	rm -f main
