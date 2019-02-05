#!/usr/bin/env bash
set -ueo pipefail
# set -x

ME=$(basename "$0")
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DOCKER_IMAGE="docker.io/digitalocean/do-agent"
VERSION=${VERSION:-$(cat target/VERSION || true)}
VERSION_REGEX="[^\\d]${VERSION}[^\\d]"
PACKAGECLOUD_DOCKER_IMAGE="docker.io/digitalocean/packagecloud:0.3.05"

FORCE_RELEASE=${FORCE_RELEASE:-0}
SUPPORTED_DEB_DISTROS="ubuntu/trusty ubuntu/utopic ubuntu/vivid ubuntu/wily ubuntu/xenial ubuntu/yakkety ubuntu/zesty ubuntu/artful ubuntu/bionic ubuntu/cosmic debian/jessie debian/stretch debian/buster"
SUPPORTED_RPM_DISTROS="fedora/27 fedora/28 fedora/29 el/6 el/7"
REMOTES=${REMOTES:-docker,packagecloud,github}
STAGE=""

CI_LOG_URL=""
[ -n "${CI_BASE_URL:-}" ] && CI_LOG_URL="${CI_BASE_URL}/tab/build/detail/${GO_PIPELINE_NAME}/${GO_PIPELINE_COUNTER}/${GO_STAGE_NAME}/${GO_STAGE_COUNTER}/${GO_JOB_NAME}"

# display usage for this script
function usage() {
	cat <<-EOF
	NAME:
	
	  $ME
	
	SYNOPSIS:

	  $ME <cmd>

	DESCRIPTION:

	    The purpose of this script is to publish build artifacts
	    to Github, docker.io, and apt/yum repositories.

	ENVIRONMENT:
		
	    VERSION (required)
	        The version to publish
	        Example: 1.0.9

	    GITHUB_AUTH_USER, GITHUB_AUTH_TOKEN (required)
	        Github access credentials

	    DOCKER_USER, DOCKER_PASSWORD  (required)
	        Docker hub access credentials

	    PACKAGECLOUD_TOKEN (required)
	        https://packagecloud.io access token

	    SLACK_WEBHOOK_URL (optional)
	        Webhook URL to send notifications. Enables Slack
	        notifications

	    REMOTES (optional)
	        Optionally only distribute to the provided
	        remotes. By default deployments will deploy
	        to the remotes supported by each deployment.
	        
	        For example: 
	          unstable deploys to docker,packagecloud
	          beta deploys to docker,packagecloud,github
	          stable deploys to docker,packagecloud,github

	COMMANDS:

	    unstable
	        Push target/ assets to packagecloud unstable.
	        Push to docker hub under the unstable and \$VERSION-rc tags.

	    beta
	        Push target/ assets to packagecloud beta.
	        Docker tag \$VERSION-rc to beta.
	        Create a github prerelease with assets.

	    stable
	        Push target/ assets to packagecloud stable.
	        Docker tag \$VERSION-rc to \$VERSION.
	        Remove prerelease flag from the github release.

	EOF
}

function main() {
	STAGE=${1:-}

	case "$STAGE" in
		unstable)
			check_version
			check_target_files
			quiet_docker_pull "${PACKAGECLOUD_DOCKER_IMAGE}"
			deploy_packagecloud "do-agent-unstable"
			docker_login && deploy_unstable_docker
			;;
		beta)
			check_version
			check_target_files
			quiet_docker_pull "${PACKAGECLOUD_DOCKER_IMAGE}"
			deploy_packagecloud "do-agent-beta"
			deploy_github_prerelease
			docker_login && promote_docker "unstable" "beta"
			;;
		stable)
			check_version
			check_target_files
			quiet_docker_pull "${PACKAGECLOUD_DOCKER_IMAGE}"
			deploy_packagecloud "do-agent"
			promote_github
			docker_login && promote_stable_docker
			;;
		help|--help|-h)
			usage
			exit 0
			;;
		*)
			abort "Unknown command '$STAGE'. See $ME --help for help"
			;;
	esac
}

# verify the VERSION env var
function check_version() {
	[[ "${VERSION:-}" =~ [0-9]+\.[0-9]+\.[0-9]+ ]] || \
		abort "VERSION is required and should be semver format (e.g. 1.2.34)"
}

function force_release_enabled() {
	if is_enabled "${FORCE_RELEASE}" ; then
		echo
		echo "WARNING! forcing a release of $VERSION"
		echo
		return 0
	fi
	return 1
}

function deploy_packagecloud() {
	repo=${1:-}
	[ -z "$repo" ] && abort "Destination repository is required. Usage: ${FUNCNAME[0]} <repo>"
	if ! remote_enabled "packagecloud"; then
		echo "packagecloud remote is disabled via REMOTES env var (${REMOTES}), skipping..."
		return
	fi

	announce "Deploying packages to packagecloud"

	if force_release_enabled ; then
		announce "Cleaning up any old versions of ${VERSION} on ${repo}"
		curl -SsL "https://${PACKAGECLOUD_TOKEN}:@packagecloud.io/api/v1/repos/digitalocean-insights/${repo}/search.json?q=${VERSION}&per_page=100" \
			| jq -r ".[] | select(.version == \"${VERSION}\") | \"\(.destroy_url)\"" \
			| tee /dev/stderr \
			| xargs -r -I{} curl -X DELETE -SsL -o /dev/null "https://${PACKAGECLOUD_TOKEN}:@packagecloud.io{}"
	else
		announce "Checking if we can deploy ${VERSION} to packagecloud ${repo}"

		status_code=$(http_status_for "https://packagecloud.io/digitalocean-insights/${repo}/packages/ubuntu/bionic/do-agent_${VERSION}_amd64.deb")
		case $status_code in
			404)
				echo "Success"
				;;
			200)
				abort "'$VERSION' has already been deployed to ${repo}. Add a new git tag or use pass FORCE_RELEASE=1."
				;;
			*)
				abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
				;;
		esac
	fi

	announce "Pushing packages to packagecloud"
	target_files | grep -P '\.(deb|rpm)$'
	echo "${SUPPORTED_DEB_DISTROS} ${SUPPORTED_RPM_DISTROS}"

	# copy the target files into a temp dir to prevent the push globs from picking up different versions
	tmpdir=.out/$(date +%s)
	mkdir -p "${tmpdir}"

	target_files | grep -P '\.(deb|rpm)$' | tee /dev/stderr | while IFS= read -r file; do
		cp -Luv "${file}" "${tmpdir}/"
	done

	for distro in ${SUPPORTED_DEB_DISTROS}; do
		package_cloud push "digitalocean-insights/${repo}/${distro}" "${tmpdir}/*.deb" &
	done

	for distro in ${SUPPORTED_RPM_DISTROS}; do
		package_cloud push "digitalocean-insights/${repo}/${distro}" "${tmpdir}/*.rpm" &
	done
	wait

	rm -rfv "${tmpdir}"

	announce "Completed deploy to packagecloud ${repo}"
}

function package_cloud() {
	docker run \
		--rm \
		-e "PACKAGECLOUD_TOKEN=${PACKAGECLOUD_TOKEN}" \
		-v "${PROJECT_ROOT}:/project" \
		-w /project \
		"${PACKAGECLOUD_DOCKER_IMAGE}" \
		"$@"
}

function check_can_deploy_github() {
	force_release_enabled && return 0
	status_code=$(http_status_for "https://api.github.com/repos/digitalocean/do-agent/releases/tags/${VERSION}")

	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been released. Add a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}


# deploy the compiled binaries and packages to github releases
function deploy_github_prerelease() {
	if ! remote_enabled "github"; then
		echo "github remote is disabled via REMOTES env var (${REMOTES}), skipping..."
		return
	fi
	check_can_deploy_github
	announce "Deploying to Github"

	create_github_release || abort "Github deploy failed"

	upload_url=$(github_asset_upload_url)

	for file in $(target_files); do
		name=$(basename "$file")

		echo "Uploading $name to github"
		github_curl \
			-X "POST" \
			-H "Content-Type: $(content_type_for "$file")" \
			--data-binary "@${file}" \
			"$upload_url?name=$name" \
			| jq -r '. | "Success: \(.name)"' &
	done
	wait
}

function check_can_promote_github() {
	force_release_enabled && return 0
	announce "Checking the state of Github release $VERSION"
	github_curl \
		--fail \
		-D /dev/stderr \
		"$(github_release_url)" \
		| jq -r '. | select(.prerelease == true) | "Found Release: \(.url)"' \
		| grep . \
		|| abort "Could not find a prerelease version $VERSION to promote. Has it already been released?"
}

function promote_github() {
	if ! remote_enabled "github"; then
		echo "github remote is disabled via REMOTES env var (${REMOTES}), skipping..."
		return
	fi
	check_can_promote_github

	announce "Removing prerelease flag from '$VERSION' on Github"
	github_curl \
		--fail \
		-i \
		-X PATCH \
		--data-binary '{"prerelease":false}' \
		"$(github_release_url)" \
		| grep 'HTTP/1.1'
}

function check_can_promote_docker() {
	force_release_enabled && return 0
	status_code=$(http_status_for "https://hub.docker.com/v2/repositories/digitalocean/do-agent/tags/${VERSION}/")
	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been released. Deploy a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}

function promote_stable_docker() {
	if ! remote_enabled "docker"; then
		echo "docker remote is disabled via REMOTES env var (${REMOTES}), skipping..."
		return
	fi
	IFS=. read -r major minor _ <<<"$VERSION"
	promote_docker "$VERSION-rc" "$VERSION"

	for tag in $major $major.$minor; do
		docker tag "$VERSION" "$tag"
		docker push "$tag"
	done
}

function promote_docker() {
	src_tag=${1:-} dest_tag=${2:-}
	[ -z "$src_tag" ] && abort "src_tag is required. Usage: ${FUNCNAME[0]} <src_tag> <dest_tag>"
	[ -z "$dest_tag" ] && abort "dest_tag is required. Usage: ${FUNCNAME[0]} <src_tag> <dest_tag>"
	if ! remote_enabled "docker"; then
		echo "docker remote is disabled via REMOTES env var (${REMOTES}), skipping..."
		return
	fi

	check_can_promote_docker
	announce "Promoting docker tag ${src_tag} to ${dest_tag}"

	quiet_docker_pull "${DOCKER_IMAGE}:$src_tag"
	docker tag "${DOCKER_IMAGE}:$src_tag" "$DOCKER_IMAGE:$dest_tag"
	docker push "$DOCKER_IMAGE:$dest_tag"
}

function http_status_for() {
	url=${1:-}
	[ -z "$url" ] && abort "Usage: ${FUNCNAME[0]} <url>"
	curl -LISsL "$url" | grep 'HTTP/1.1' | awk '{ print $2 }'
}

# print the content type header for the provided file
function content_type_for() {
	file=${1:-}
	[ -z "$file" ] && abort "Usage: ${FUNCNAME[0]} <file>"
	case $file in
		*.deb) echo "application/vnd.debian.binary-package" ;;
		*.rpm) echo "application/x-rpm" ;;
		*.tar.gz) echo "application/gzip" ;;
		*) echo "application/octet-stream"
	esac
}

# get the asset upload URL for VERSION
function github_asset_upload_url() {
	if base=$(github_release_url); then
		echo "${base/api/uploads}/assets"
	else
		return 1
	fi
}

# get the base release url for VERSION
function github_release_url() {
	github_curl \
		--fail \
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '.url' \
		| grep .
}


function rm_old_assets() {
	assets=$(github_curl \
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '.assets[].url')
	for asset in $assets; do
		echo "Removing old asset $asset"
		github_curl \
			-X DELETE \
			"$asset" &
		wait
	done
}

# create a github release for VERSION
function create_github_release() {
	announce "Checking for existing Github release"
	if github_release_url >/dev/null; then
		echo "Github release exists $VERSION"
		# we cannot upload the same asset twice so we have to delete
		# the old assets before we can commense with uploads
		rm_old_assets || abort "failed to purge Github release assets"
		return 0
	fi

	announce "Creating Github release $VERSION"

	data=$(cat <<-EOF
	{ "tag_name": "$VERSION", "name": "$VERSION", "prerelease": true, "target_commitish": "master" }
	EOF
	)
	echo "POST: $data"
	github_curl \
		-o /dev/null \
		-X POST \
		-H 'Content-Type: application/json' \
		-d "$data" \
		https://api.github.com/repos/digitalocean/do-agent/releases
}

function docker_login() {
	# gocd has an old version of docker that does not have --pasword-stdin
	docker login -u "$DOCKER_USER" -p "$DOCKER_PASSWORD"
}

function check_can_deploy_docker() {
	force_release_enabled && return 0
	status_code=$(http_status_for "https://hub.docker.com/v2/repositories/digitalocean/do-agent/tags/${VERSION}-rc/")
	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION-rc' has already been released. Add a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}

# build and push the RC docker hub image. This image is considered unstable
# and should only be used for testing purposes
function deploy_unstable_docker() {
	check_can_deploy_docker
	if ! remote_enabled "docker"; then
		echo "docker remote is disabled via REMOTES env var (${REMOTES}), skipping..."
		return
	fi
	announce "Pushing docker images"

	for tag in unstable ${VERSION}-rc; do
		docker build . -t "$DOCKER_IMAGE:${tag}"
		docker push "$DOCKER_IMAGE:${tag}"
	done
}

# list the artifacts within the target/ directory
function target_files() {
	find target/pkg -type f -iname "*${VERSION_REGEX}*" \
		| grep .
}

function check_target_files() {
	target_files > /dev/null || abort "No packages for $VERSION were found in target/.  Did you forget to run make?"
}

function quiet_docker_pull() {
	img=${1:-}
	[ -z "$img" ] && abort "img param is required. Usage: ${FUNCNAME[0]} <img>"
	docker pull "${img}" | grep -e 'Pulling from' -e Digest -e Status -e Error
}

# call CURL with github authentication
function github_curl() {
	# if user and token are empty then bash will exit because of unbound vars
	curl -SsL \
		--fail \
		-u "${GITHUB_AUTH_USER}:${GITHUB_AUTH_TOKEN}" \
		"$@"
}

# abort with an error message
function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file:$func:$line: $1" > /dev/stderr
	exit 1
}

# print something to STDOUT with formatting
# Usage: announce "Some message"
#
# Examples:
#    announce "Begin execution of something"
#    announce "All is well"
#
function announce() {
	msg=${1:-}
	[ -z "$msg" ] && abort "Usage: ${FUNCNAME[0]} <msg>"
	echo ":::::::::::::::::::::::::::::::::::::::::::::::::: $msg ::::::::::::::::::::::::::::::::::::::::::::::::::" > /dev/stderr
}

function is_enabled() {
	v=$(echo "${1:-}" | tr '[:upper:]' '[:lower:]')
	if [[ "${v}" =~ ^y(es)?|t(rue)?|1$ ]]; then
		return 0
	else
		return 1
	fi
}

function remote_enabled() {
	remote=${1:-}
	[ -z "$remote" ] && abort "remote param is required. Usage: ${FUNCNAME[0]} <remote>"
	[[ "$REMOTES" =~ $remote ]]
}


# send a slack notification or fallback to STDERR
# Usage: notify <success> <msg> [link]
#
# Examples:
#    notify 0 "Deployed to Github failed!"
#    notify "true" "Success!" "https://github.com/"
#
function notify() {
	success=${1:-} msg=${2:-} link=${3:-}

	[ -z "${SLACK_WEBHOOK_URL:-}" ] && return 0

	color="green"
	[[ "$success" =~ ^(false|0|no)$ ]] && color="red"

	payload=$(cat <<-EOF
	{
	  "attachments": [
	    {
	      "fallback": "${msg}",
	      "color": "${color}",
	      "title": "${msg}",
	      "title_link": "${link}",
	      "text": "${msg}",
	      "fields": [
		{
		  "title": "App",
		  "value": "do-agent",
		  "short": true
		},
		{
		  "title": "Version",
		  "value": "${VERSION}",
		  "short": true
		},
		{
		  "title": "Stage",
		  "value": "${STAGE}",
		  "short": true
		},
		{
		  "title": "User",
		  "value": "${USER}",
		  "short": true
		},
		{
		  "title": "Source",
		  "value": "$(hostname -s)",
		  "short": true
		}
	      ]
	    }
	  ]
	}
	EOF
	)

	curl -sS -X POST \
		--fail \
		--data-binary "$payload" \
		"${SLACK_WEBHOOK_URL}" > /dev/null

	# always pass to prevent pipefailures
	return 0
}

function notify_exit() {
	if [ "$1" != "0" ]; then
		notify 0 "Deploy failed" "${CI_LOG_URL:-}"
	else
		notify 1 "Deploy succeeded"
	fi
}
trap 'notify_exit $?' ERR EXIT INT TERM

main "$@"
