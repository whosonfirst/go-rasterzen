CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-cache-s3
	cp *.go src/github.com/whosonfirst/go-whosonfirst-cache-s3
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

# we move the aws stuff in to the root because we need to test AWS errors locally

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-cache"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-aws/..."
	mv src/github.com/whosonfirst/go-whosonfirst-aws/vendor/github.com/aws src/github.com/

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt cmd/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/s3-cache-get cmd/s3-cache-get.go
	@GOPATH=$(GOPATH) go build -o bin/s3-cache-put cmd/s3-cache-put.go