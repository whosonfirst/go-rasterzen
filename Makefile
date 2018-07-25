CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src; then mkdir src; fi
	mkdir -p src/github.com/whosonfirst/go-rasterzen
	cp *.go src/github.com/whosonfirst/go-rasterzen/
	cp -r nextzen src/github.com/whosonfirst/go-rasterzen/
	cp -r tile src/github.com/whosonfirst/go-rasterzen/
	cp -r http src/github.com/whosonfirst/go-rasterzen/
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
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/geojson2svg/pkg/geojson2svg"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-cache-s3"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-cli"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/algnhsa"
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
	go fmt tile/*.go
	go fmt server/*.go
	go fmt nextzen/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/rasterd cmd/rasterd.go

lambda:	
	@make self
	if test -f main; then rm -f main; fi
	if test -f deployment.zip; then rm -f deployment.zip; fi
	@GOPATH=$(GOPATH) GOOS=linux go build -o main cmd/rasterd.go
	zip deployment.zip main
	rm -f main
