mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GIT_HASH:=$(shell git rev-parse --short HEAD)


build:
	go build -o ./bin/asciinema-edit -ldflags "-s -w" ./cmd/main.go

fmt:
	go fmt ./...

test:
	go test ./... -v

release: VERSION := $(shell cat ./VERSION)
release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist


test-coverage:
	go test -timeout 30s -coverprofile $(ROOT_DIR)/cover.out  $(ROOT_DIR)/...


.PHONY: build
