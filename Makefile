APP_NAME := kbom
GCR_ORG := ksoc-public
GITHUB_ORG := rad-security
GIT_REPO ?= github.com/$(GITHUB_ORG)/$(APP_NAME)
VERSION := $(shell SEP="-" bash scripts/version)

BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')
LAST_COMMIT_USER ?= $(shell git log -1 --format='%cn <%ce>')
LAST_COMMIT_HASH ?= $(shell git log -1 --format=%H)
LAST_COMMIT_TIME ?= $(shell git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S')

export APP_NAME
export GCR_ORG
export GITHUB_ORG
export VERSION
export BUILD_TIME
export LAST_COMMIT_USER
export LAST_COMMIT_HASH
export LAST_COMMIT_TIME

.PHONY: initialise
initialise: ## Initialises the project, set ups git hooks
	pre-commit install

.PHONY: release
release: ## Builds a release
	goreleaser release --clean --timeout 90m

.PHONY: semtag-%
semtag-%: ## Creates a new tag using semtag
	semtag final -s $*

.PHONY: snapshot
snapshot: ## Builds a snapshot release
	GORELEASER_CURRENT_TAG=$(GORELEASER_CURRENT_TAG) \
		goreleaser build --snapshot --clean --single-target --timeout 90m

.PHONY: docker_push_all
docker_push_all: ## Pushes all docker images
	docker push $$(docker images -a  | grep $(APP_NAME) | awk '{ print $$1 ":" $$2 }')

.PHONY: build
build: ## Builds kbom binary
	CGO_ENABLED=0 \
	go build \
	-v \
	-ldflags "-s -w \
	-X '$(GIT_REPO)/internal/config.AppName=$(APP_NAME)' \
	-X '$(GIT_REPO)/internal/config.AppVersion=$(VERSION)' \
	-X '$(GIT_REPO)/internal/config.BuildTime=$(BUILD_TIME)' \
	-X '$(GIT_REPO)/internal/config.LastCommitUser=$(LAST_COMMIT_USER)' \
	-X '$(GIT_REPO)/internal/config.LastCommitHash=$(LAST_COMMIT_HASH)' \
	-X '$(GIT_REPO)/internal/config.LastCommitTime=$(LAST_COMMIT_TIME)'" \
	-o $(APP_NAME) .

.PHONY: test
test: ## Runs unit tests
	go test -coverprofile coverage.out -v --race ./...
	go tool cover -html=coverage.out -o coverage_report.html

.PHONY: help
help: ## Displays this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ \
	{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
