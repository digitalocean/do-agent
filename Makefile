GOOS   ?= linux
GOARCH ?= amd64

ifeq ($(GOARCH),386)
PKG_ARCH = i386
else
PKG_ARCH = amd64
endif

############
## macros ##
############

mkdir        = @mkdir -p $(dir $@)
cp           = @cp $< $@
print        = @printf "\n:::::::::::::::: [$(shell date -u)] $@ ::::::::::::::::\n"
touch        = @touch $@
jq           = @docker run --rm -i colstrom/jq
shellcheck   = @docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) koalaman/shellcheck:v0.6.0
revgrep      = docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) digitalocean/revgrep:latest
fpm          = @docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) digitalocean/fpm:latest
vault        = @docker run --rm -i -u $(shell id -u) --net=host -e "VAULT_TOKEN=$(shell cat .vault-token || echo)" docker.internal.digitalocean.com/eng-insights/vault:0.11.5
now          = $(shell date -u)
git_rev      = $(shell git rev-parse --short HEAD)
git_tag      = $(subst v,,$(shell git describe --tags --abbrev=0))
VERSION     ?= $(git_tag)

dockerlogin = @echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

linter = docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -e "GO111MODULE=on" -e "GOFLAGS=-mod=vendor" -e "GOCACHE=$(CURDIR)/target/.cache/go" \
	-u $(shell id -u) golangci/golangci-lint:v1.16 \
	golangci-lint run --no-config --disable-all -E gosec -E interfacer -E vet -E deadcode -E gocyclo -E golint \
	-E varcheck -E dupl -E ineffassign -E misspell -E unconvert -E gosec -E nakedret -E goconst -E gofmt -E unparam \
	-E prealloc \
	./...

# where to mount the CURDIR in docker
docker_dir = /home/do-agent

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
	golang:1.16.3 \
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

###########
## paths ##
###########

out             := target
package_dir     := $(out)/pkg
cache           := $(out)/.cache
project         := $(notdir $(CURDIR))# project name
pkg_project     := $(subst _,-,$(project))# package cannot have underscores in the name
gofiles         := $(shell find -type f -iname '*.go' ! -path './vendor/*')
vendorgofiles   := $(shell find -type f -iname '*.go' -path './vendor/*')
shellscripts    := $(shell find -type f -iname '*.sh' ! -path './repos/*' ! -path './vendor/*' ! -path './.git/*')
# the name of the binary built with local resources
binary          := $(out)/$(project)-$(GOOS)-$(GOARCH)
cover_profile   := $(out)/.coverprofile

# output packages
# deb files should end with _version_arch.deb
# rpm files should end with -version-release.arch.rpm
base_package := $(package_dir)/$(pkg_project).$(VERSION).$(PKG_ARCH).BASE.deb
deb_package  := $(package_dir)/$(pkg_project)_$(VERSION)_$(PKG_ARCH).deb
rpm_package  := $(package_dir)/$(pkg_project).$(VERSION).$(PKG_ARCH).rpm
tar_package  := $(package_dir)/$(pkg_project).$(VERSION).tar.gz

#############
## targets ##
#############

build: $(binary)
$(binary): $(gofiles) $(vendorgofiles)
	$(print)
	$(mkdir)
	$(go) build -ldflags $(ldflags) -o "$@" ./cmd/$(project)

package: release
release: target/VERSION
	$(print)
	@GOOS=linux GOARCH=386 $(MAKE) build deb rpm tar
	@GOOS=linux GOARCH=amd64 $(MAKE) build deb rpm tar

lint: $(cache)/lint $(cache)/shellcheck
$(cache)/lint: $(gofiles)
	$(print)
	$(mkdir)
	@$(linter) ./... | $(revgrep) master
	$(touch)

shellcheck: $(cache)/shellcheck
$(cache)/shellcheck: $(shellscripts)
	$(print)
	$(mkdir)
	@$(shellcheck) --version
	@$(shellcheck) $^
	$(touch)

test: $(cover_profile)
$(cover_profile): $(gofiles)
	$(print)
	$(mkdir)
	@$(go) test -coverprofile=$@ ./...

clean:
	$(print)
	@rm -rf $(out)
.PHONY: clean

ci: login clean test package lint shellcheck
.PHONY: ci

login:
	$(print)
	@$(dockerlogin)
.PHONY: login

.PHONY: target/VERSION
target/VERSION:
	$(print)
	$(mkdir)
	@echo $(VERSION) > $@

# used to create a base package with common functionality
$(base_package): $(binary)
	$(print)
	$(mkdir)
	@$(fpm) --output-type deb \
		--verbose \
		--input-type dir \
		--force \
		--architecture $(PKG_ARCH) \
		--package $@ \
		--no-depends \
		--name $(pkg_project) \
		--maintainer "DigitalOcean" \
		--version $(VERSION) \
		--description "DigitalOcean stats collector" \
		--license apache-2.0 \
		--vendor DigitalOcean \
		--url https://github.com/digitalocean/do-agent \
		--log info \
		--after-install packaging/scripts/after_install.sh \
		--after-remove packaging/scripts/after_remove.sh \
		--config-files /etc/init/do-agent.conf \
		--config-files /etc/systemd/system/do-agent.service \
		$<=/opt/digitalocean/bin/do-agent \
		scripts/update.sh=/opt/digitalocean/do-agent/scripts/update.sh \
		packaging/etc/init/do-agent.conf=/etc/init/do-agent.conf \
		packaging/etc/systemd/system/do-agent.service=/etc/systemd/system/do-agent.service
.INTERMEDIATE: $(base_package)

deb: $(deb_package)
$(deb_package): $(base_package)
	$(print)
	$(mkdir)
	@$(fpm) --output-type deb \
		--verbose \
		--input-type deb \
		--force \
		--depends cron \
		--deb-group nobody \
		--deb-user do-agent \
		-p $@ \
		$<
	chown -R $(USER):$(USER) target
# print information about the compiled deb package
	@docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" ubuntu:xenial /bin/bash -c 'dpkg --info $@ && dpkg -c $@'

.PHONY: install_deb
install_deb: deb
	@dpkg -i $(deb_package)

rpm: $(rpm_package)
$(rpm_package): $(base_package)
	$(print)
	$(mkdir)
	@$(fpm) \
		--verbose \
		--output-type rpm \
		--input-type deb \
		--depends cronie \
		--rpm-group nobody \
		--rpm-user do-agent \
		--rpm-posttrans packaging/scripts/after_install.sh \
		--force \
		-p $@ \
		$<
	chown -R $(USER):$(USER) target
# print information about the compiled rpm package
	@docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" centos:7 rpm -qilp $@

tar: $(tar_package)
$(tar_package): $(base_package)
	$(print)
	$(mkdir)
	@$(fpm) \
		--verbose \
		--output-type tar \
		--input-type deb \
		--force \
		-p $@ \
		$<
	chown -R $(USER):$(USER) target
# print all files within the archive
	@docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" ubuntu:xenial tar -ztvf $@

.INTERMEDIATE: .vault-token
.vault-token:
	$(print)
	$(vault) write -field token auth/approle/login role_id=$(VAULT_ROLE_ID) secret_id=$(VAULT_SECRET_ID) \
		| cp /dev/stdin $@

.INTERMEDIATE: .id_rsa
.id_rsa: .vault-token
	$(print)
	$(vault) read --field ssh-priv-key secret/agent/packager/terraform \
		| cp /dev/stdin $@

.INTERMEDIATE: sonar-agent.key
sonar-agent.key: .vault-token
	$(print)
	$(mkdir)
	$(vault) read --field gpg secret/agent/packager/key > $@

.PHONY: deploy
deploy: .id_rsa sonar-agent.key
ifndef release
	$(error Usage: make deploy release=(unstable|beta|stable))
endif
	@RSYNC_KEY_FILE=$(CURDIR)/.id_rsa GPG_PRIVATE_KEY_FILE=$(CURDIR)/sonar-agent.key ./scripts/deploy.sh $(release)

