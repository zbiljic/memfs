SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
THIS_OS := $(shell uname)

default: help

# Only for Travis CI compliance
.PHONY: bootstrap
bootstrap: lint-deps # Install all dependencies

.PHONY: lint-deps
lint-deps: ## Install linter dependencies
	@echo "==> Updating linter dependencies..."
	@go get -u github.com/alecthomas/gometalinter
	@gometalinter --install && echo "Installed gometalinter"

.PHONY: check
check: ## Lint the source code
	@echo "==> Linting source code..."
	@gometalinter \
		--deadline 10m \
		--vendor \
		--exclude='.*\.generated\.go' \
		--exclude='.*bindata_assetfs\.go' \
		--skip="ui/" \
		--sort="path" \
		--aggregate \
		--enable-gc \
		--disable-all \
		--enable goimports \
		--enable misspell \
		--enable vet \
		--enable deadcode \
		--enable varcheck \
		--enable ineffassign \
		--enable structcheck \
		--enable unconvert \
		--enable gas \
		--enable gofmt \
		./...

.PHONY: checkscripts
checkscripts: ## Lint shell scripts
	@echo "==> Linting scripts..."
	@shellcheck ./scripts/*.sh

.PHONY: test
test: LOCAL_PACKAGES = $(shell go list ./... | grep -v '/vendor/')
test: ## Run the test suite and/or any other tests
	@echo "==> Running test suites..."
	@go test \
		-cover \
		-timeout=900s \
		$(LOCAL_PACKAGES)

.PHONY: coverage
coverage: ## Create coverage report
	@echo "==> Running all coverage..."
	@(env bash scripts/go-coverage.sh)

.PHONY: buildchecks
buildchecks: GOPATH=$(shell go env GOPATH)
buildchecks: ## Pre-build checks
	@echo "==> Running pre-build checks..."
	@echo "Checking project is in GOPATH"
	@(env bash scripts/checkgopath.sh)

.PHONY: clean
clean: GOPATH=$(shell go env GOPATH)
clean: ## Remove build artifacts
	@echo "==> Cleaning build artifacts..."
	@rm -fv coverage.txt
	@find . -name '*.test' | xargs rm -fv
	@rm -rf "$(PROJECT_ROOT)/memfs"
	@rm -rf "$(PROJECT_ROOT)/release/"
	@rm -f "$(GOPATH)/bin/memfs"

HELP_FORMAT="    \033[36m%-15s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo

FORCE:
