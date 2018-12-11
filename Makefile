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

find         = $(shell find . -name \*.$1 -type f ! -path '*/vendor/*' ! -path '*/target/*')
mkdir        = @mkdir -p $(dir $@)
cp           = @cp $< $@
print        = @printf "\n:::::::::::::::: [$(shell date -u)] $@ ::::::::::::::::\n"
touch        = @touch $@
jq           = @docker run --rm -i colstrom/jq
shellcheck   = @docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) koalaman/shellcheck:v0.6.0
gometalinter = @docker run --rm -i -v "$(CURDIR):/go/src/$(importpath)" -w "/go/src/$(importpath)" -u $(shell id -u) digitalocean/gometalinter:2.0.11
fpm          = @docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) digitalocean/fpm:latest

go    = docker run --rm -i \
	-u "$(shell id -u)" \
	-e "GOOS=$(GOOS)" \
	-e "GOARCH=$(GOARCH)" \
	-e "GOPATH=/gopath" \
	-e "GOCACHE=/gopath/src/$(importpath)/target/.cache/go" \
	-v "$(CURDIR):/gopath/src/$(importpath)" \
	-w "/gopath/src/$(importpath)" \
	golang:1.11.2 \
	go

now          = $(shell date -u +"%F %T %Z")
git_revision = $(shell git rev-parse --short HEAD)
git_branch   = $(shell git rev-parse --abbrev-ref HEAD)
git_tag      = $(shell git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')
# git_tag produces v0.0.0-0-g<SHA>. Which is <tag>-<release>-g<SHA> where tag
# is the most recent tag, release is the amount of commits since the latest
# tag, and SHA is the SHA of the most recent commit prefixed with a 'g'. We
# need the format to be 1.0.0-1 so we strip the -g<revision> and the v prefix
pkg_version = $(subst -g$(git_revision),,$(subst v,,$(shell git describe --tags)))
ldflags     = '\
	-X "main.version=$(git_tag)" \
	-X "main.revision=$(git_revision)" \
	-X "main.branch=$(git_branch)" \
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
importpath      := github.com/digitalocean/$(project)# import path used in gocode
gofiles         := $(shell find -type f -iname '*.go' ! -path './vendor/*')
vendorgofiles   := $(shell find -type f -iname '*.go' -path './vendor/*')
shellscripts    := $(shell find -type f -iname '*.sh' ! -path './repos/*' ! -path './vendor/*')

# the name of the binary built with local resources
binary             := $(out)/$(project)_$(GOOS)_$(GOARCH)
cover_profile      := $(out)/.coverprofile

# output packages
# deb files should end with _version_arch.deb
# rpm files should end with -version-release.arch.rpm
base_package := $(package_dir)/$(pkg_project).$(pkg_version).$(PKG_ARCH).BASE.deb
deb_package  := $(package_dir)/$(pkg_project)_$(pkg_version)_$(PKG_ARCH).deb
rpm_package  := $(package_dir)/$(pkg_project).$(pkg_version).$(PKG_ARCH).rpm
tar_package  := $(package_dir)/$(pkg_project).$(pkg_version).tar.gz

# use the binary's mtime for epoch for consistency. This needs to be lazily
# evaluated since the binary does not yet exist
epoch = $(shell date '+%s' -r $(binary))

#############
## targets ##
#############

build: $(binary)
$(binary): $(gofiles) $(vendorgofiles)
	$(print)
	$(mkdir)
	$(go) build \
	     -ldflags $(ldflags) \
	     -o "$@" \
	     ./cmd/$(project)

package: release
release: target/VERSION
	$(print)
	@GOOS=linux GOARCH=386 $(MAKE) build deb rpm tar
	@GOOS=linux GOARCH=amd64 $(MAKE) build deb rpm tar

lint: $(cache)/lint $(cache)/shellcheck
$(cache)/lint: $(gofiles)
	$(print)
	$(mkdir)
	@$(gometalinter) --config=gometalinter.json ./...
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

ci: clean lint shellcheck test package
.PHONY: ci

.PHONY: target/VERSION
target/VERSION:
	$(print)
	$(mkdir)
	@echo $(git_tag) > $@

# used to create a base package with common functionality
$(base_package): $(binary)
	$(print)
	$(mkdir)
	@$(fpm) --output-type deb \
		--verbose \
		--input-type dir \
		--epoch $(epoch) \
		--force \
		--architecture $(PKG_ARCH) \
		--package $@ \
		--no-depends \
		--name $(pkg_project) \
		--maintainer "DigitalOcean" \
		--version $(pkg_version) \
		--description "DigitalOcean stats collector" \
		--license apache-2.0 \
		--vendor DigitalOcean \
		--url https://github.com/digitalocean/do-agent \
		--log info \
		--after-install packaging/scripts/after_install.sh \
		--after-remove packaging/scripts/after_remove.sh \
		$<=/usr/local/bin/do-agent \
		scripts/update.sh=/opt/digitalocean/do-agent/scripts/update.sh
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
		--conflicts do-agent \
		--replaces do-agent \
		--deb-group nobody \
		--deb-user nogroup \
		-p $@ \
		$<
	chown -R $(USER):$(USER) target
# print information about the compiled deb package
	@docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" ubuntu:xenial /bin/bash -c 'dpkg --info $@ && dpkg -c $@'


rpm: $(rpm_package)
$(rpm_package): $(base_package)
	$(print)
	$(mkdir)
	@$(fpm) \
		--verbose \
		--output-type rpm \
		--input-type deb \
		--depends cronie \
		--conflicts do-agent \
		--replaces do-agent \
		--rpm-group nobody \
		--rpm-user nobody \
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


.vault-token:
	$(print)
	@docker run -u $(shell id -u) --net=host --rm -i \
		docker.internal.digitalocean.com/eng-insights/vault:0.11.5 \
		write -field token auth/approle/login role_id=$(VAULT_ROLE_ID) secret_id=$(VAULT_SECRET_ID) \
		| cp /dev/stdin $@
.INTERMEDIATE: .vault-token

sonar-agent.key: .vault-token
	$(print)
	@docker run -u $(shell id -u) --net=host --rm -i -e "VAULT_TOKEN=$(shell cat $^)" \
		docker.internal.digitalocean.com/eng-insights/vault:0.11.5 \
		read --field gpg secret/agent/packager/key \
		| cp /dev/stdin $@

.PHONY: deploy
deploy:
	./scripts/deploy.sh all

.PHONY: promote
promote:
	./scripts/deploy.sh promote
