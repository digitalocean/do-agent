.DEFAULT_GOAL := ci

GOOS       ?= linux
GOARCH     ?= amd64

out           := target
package_dir   := $(out)/pkg
cache         := $(out)/.cache
docker_dir    := /home/do-agent
shellscripts  := $(shell find -type f -iname '*.sh' ! -path './repos/*' ! -path './vendor/*' ! -path './.git/*')
go_version    := $(shell sed -En 's/^go[[:space:]]+([[:digit:].]+)$$/\1/p' go.mod)
print         = @printf "\n:::::::::::::::: [$(shell date -u)] $@ ::::::::::::::::\n"

go = \
	docker run --rm -i \
	-u "$(shell id -u):$(shell id -g)" \
	-e "GOOS=$(GOOS)" \
	-e "GOARCH=$(GOARCH)" \
	-e "GO111MODULE=on" \
	-e "GOFLAGS=-mod=vendor" \
	-e "GOCACHE=$(docker_dir)/target/.cache/go" \
	-v "$(CURDIR):$(docker_dir)" \
	-w "$(docker_dir)" \
	golang:$(go_version) \
	go

shellcheck = \
	docker run --rm -i \
	-u "$(shell id -u):$(shell id -g)" \
	-v "$(CURDIR):$(docker_dir)" \
	-w "$(docker_dir)" \
	koalaman/shellcheck:v0.6.0

linter = \
	docker run --rm -i \
	-w "$(docker_dir)" \
	-e "GO111MODULE=on" \
	-e "GOFLAGS=-mod=vendor" \
	-v "$(CURDIR):$(docker_dir)" \
	golangci/golangci-lint:v1.50.1

clean:
	$(print)
	rm -rf ./target

test:
	$(print)
	$(go) test -v ./...

build:
	$(print)
	$(go) build -v ./...

shell:
	$(print)
	$(shellcheck) $(shellscripts)

lint:
	$(print)
	$(linter) golangci-lint run ./...

ci: clean build test lint shell
