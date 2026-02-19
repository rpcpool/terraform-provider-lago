SHELL := /usr/bin/env bash

BINARY := terraform-provider-lago

.PHONY: build test testacc fmt vet install clean

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

testacc:
	LAGO_ACC=1 go test ./internal/provider -v -count=1

fmt:
	gofmt -w $$(find . -name '*.go')

vet:
	go vet ./...

install:
	go install .

clean:
	rm -rf bin
