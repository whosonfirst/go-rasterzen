FROM golang:1.12-alpine as builder

RUN mkdir /build

COPY . /build/go-rasterzen

RUN apk update && apk upgrade \
    && apk add make libc-dev gcc \
    && cd /build/go-rasterzen \
    && make tools

FROM alpine:latest

COPY --from=builder /build/go-rasterzen/bin/rasterzen-seed /usr/local/bin/rasterzen-seed

RUN apk update && apk upgrade \
    && apk add ca-certificates