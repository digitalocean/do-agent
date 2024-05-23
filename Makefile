.DEFAULT_GOAL := ci

GOOS       ?= linux
GOARCH     ?= amd64

############
## macros ##
############
now          = $(shell date -u)
go_version   = $(shell sed -En 's/^go[[:space:]]+([[:digit:].]+)$$/\1/p' go.mod)
git_rev      = $(shell git rev-parse --short HEAD)
git_tag      = $(subst v,,$(shell git describe --tags --abbrev=0))
VERSION     ?= $(git_tag)
print        = @printf "\n:::::::::::::::: [$(shell date -u) | $(VERSION)] $@ ::::::::::::::::\n"

###########
## paths ##
###########
out           := target
package_dir   := $(out)/pkg
cache         := $(out)/.cache
docker_dir    := /home/do-agent
project       := $(notdir $(CURDIR))#project name
binary        := $(out)/$(project)-$(GOOS)-$(GOARCH)
gofiles       := $(shell find -type f -iname '*.go' ! -path './vendor/*')
shellscripts  := $(shell find -type f -iname '*.sh' ! -path './repos/*' ! -path './vendor/*' ! -path './.git/*')
vendorgofiles := $(shell find -type f -iname '*.go' -path './vendor/*')

ifneq ($(DOCKER_BUILD),1)
go = docker run --rm -i \
	-u "$(shell id -u)" \
	-e "GOOS=$(GOOS)" \
	-e "GOARCH=$(GOARCH)" \
	-e "GO111MODULE=on" \
	-e "GOFLAGS=-mod=vendor" \
	-e "GOCACHE=$(docker_dir)/target/.cache/go" \
	-v "$(CURDIR):$(docker_dir)" \
	-w "$(docker_dir)" \
	golang:$(go_version) \
	go
else
go = GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	GO111MODULE=on \
	GOFLAGS=-mod=vendor \
	GOCACHE=$(docker_dir)/target/.cache/go \
	$(shell which go)
endif

ldflags = '\
	-s -w \
	-X "main.version=$(VERSION)" \
	-X "main.revision=$(git_rev)" \
	-X "main.buildDate=$(now)" \
'

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
	golangci/golangci-lint:v1.58.2

#############
## targets ##
#############
clean:
	$(print)
	rm -rf ./target

test:
	$(print)
	$(go) test -v ./...

build: $(binary)
$(binary): $(gofiles) $(vendorgofiles)
	$(print)
	$(go) build -buildvcs=false -ldflags $(ldflags) -o "$(docker_dir)/$@" ./cmd/$(project)

shell:
	$(print)
	$(shellcheck) $(shellscripts)

lint:
	$(print)
	$(linter) golangci-lint run ./...

ci: clean build test lint shell
