.PHONY: all test clean build dependencies

CONFIG_PATH=github.com/digitalocean/do-agent/config

CURRENT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
CURRENT_HASH=$(shell git rev-parse --short HEAD)

ifeq ("$(shell git name-rev --tags --name-only $(shell git rev-parse HEAD))", "undefined")
	RELEASE=dev
else
	RELEASE=$(shell git name-rev --tags --name-only $(shell git rev-parse HEAD) | sed 's/\^.*$///')
endif

LAST_RELEASE=$(shell git describe --tags $(shell git rev-list --tags --max-count=1))
GOFLAGS=-ldflags "-X $(CONFIG_PATH).build=$(CURRENT_BRANCH).$(CURRENT_HASH) -X $(CONFIG_PATH).version=$(RELEASE)"

GOVENDOR=$(GOPATH)/bin/govendor

all: build test

build: dependencies
	@echo ">> build version=$(RELEASE)"
	@echo ">> Building system native"
	@go build $(GOFLAGS) -o do-agent cmd/do-agent/main.go
	@echo ">> Creating build directory"
	@mkdir -p build
	@echo ">> Building linux 386"
	@env GOOS=linux GOARCH=386 go build $(GOFLAGS) -o build/do-agent_linux_386 cmd/do-agent/main.go
	@echo ">> Building linux amd64"
	@env GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o build/do-agent_linux_amd64 cmd/do-agent/main.go

build-latest-release: checkout-latest-release build

checkout-latest-release: master-branch-check
	git fetch --tags
	git checkout $(LAST_RELEASE)

install:
	@go get $(GOFLAGS) ./...

test: dependencies
	@echo " ==Running go test=="
	@go test -v $(shell go list ./... | grep -v /vendor/)
	@echo " ==Running go vet=="
	@go vet $(shell go list ./... | grep -v /vendor/)
	@go get -u github.com/golang/lint/golint
	@echo " ==Running golint=="
	@golint ./... | grep -v '^vendor\/' | grep -v ".pb.*.go:" || true
	@echo " ==Done testing=="

clean:
	rm do-agent
	rm -fr build

dependencies: $(GOVENDOR)
	@echo ">> fetching dependencies"
	@$(GOVENDOR) sync

$(GOVENDOR):
	@echo ">> fetching govendor"
	@go get -u github.com/kardianos/govendor

list-latest-release:
	@echo $(LAST_RELEASE)

release-major-version: master-branch-check
	@echo ">> release major version"
	$(eval RELEASE_VERSION=$(shell echo $(LAST_RELEASE) | awk '{split($$0,a,"."); print a[1]+1"."0"."0}'))
	@echo "Updating release version from=$(LAST_RELEASE) to=$(RELEASE_VERSION)"
	git tag $(RELEASE_VERSION) -m"make release-major-version $(RELEASE_VERSION)"
	git push origin --tags

release-minor-version: master-branch-check
	@echo "release minor version"
	$(eval RELEASE_VERSION=$(shell echo $(LAST_RELEASE) | awk '{split($$0,a,"."); print a[1]"."a[2]+1"."0}'))
	@echo "Updating release version from=$(LAST_RELEASE) to=$(RELEASE_VERSION)"
	git tag $(RELEASE_VERSION) -m"make release-minor-version $(RELEASE_VERSION)"
	git push origin --tags

release-patch-version: master-branch-check
	@echo "release patch version"
	$(eval RELEASE_VERSION=$(shell echo $(LAST_RELEASE) | awk '{split($$0,a,"."); print a[1]"."a[2]"."a[3]+1}'))
	@echo "Updating release version from=$(LAST_RELEASE) to=$(RELEASE_VERSION)"
	git tag $(RELEASE_VERSION) -m"make release-patch-version $(RELEASE_VERSION)"
	git push origin --tags

master-branch-check:
ifeq ("$(shell git rev-parse --abbrev-ref HEAD)", "master")
	@echo "Current branch is master"
else
	$(error Action requires the master branch)
endif
