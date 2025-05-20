SHELL:=/bin/bash

COLOR_RED := \e[1;31m
COLOR_GRN := \e[1;32m
COLOR_YEL := \e[1;33m
COLOR_BLU := \e[1;34m
COLOR_MAG := \e[1;35m
COLOR_CYN := \e[1;36m
COLOR_END := \e[0m

@print = @printf "${2}${1}${COLOR_END}\n"
print = printf "${2}${1}${COLOR_END}\n"

## Current path
PATH_CURRENT = $(shell pwd)

##? The path where the binary file will be placed
PATH_OUTPUT_BIN ?= $(PATH_CURRENT)/.bin

##? File name for code-coverage
GO_TEST_COVERAGE_FILE_NAME ?= coverage.out

all: help

.PHONY: install build dependency generate config tools-lint-ci tools-lint tools-format tools-test
##@ Build
# gcflags and asmflags (trimpath)- remove prefix from recorded source file paths
# ldflags forward the version value to a variable Version
install: dependency generate ## Build binary files
	$(call @print,Build binaries to GOPATH,${COLOR_BLU})
	@ERR=0; \
	go build -gcflags="-trimpath=$(PATH_CURRENT)" \
			-asmflags="-trimpath=$(PATH_CURRENT)" \
			-ldflags "-X main.Version=$(APP_VERSION)" \
			-o "$${GOPATH}/bin/$${BIN}" || { \
		ERR=$$?; \
	}; \
	if [ $$ERR != 0 ]; then \
		exit $$ERR; \
	fi

build: dependency generate config ## Build binary files
	$(call @print,Build binaries to .bin,${COLOR_BLU})
	@mkdir -p $(PATH_OUTPUT_BIN)
	@ERR=0; \
	go build -gcflags="-trimpath=$(PATH_CURRENT)" \
			-asmflags="-trimpath=$(PATH_CURRENT)" \
			-ldflags "-X main.Version=$(APP_VERSION)" \
			-o "$(PATH_OUTPUT_BIN)/$${BIN}" || { \
		ERR=$$?; \
	}; \
	if [ $$ERR != 0 ]; then \
		exit $$ERR; \
	fi

config: ## Cope and replace configs
	@mkdir -p $(PATH_OUTPUT_BIN)
	@CONFIG="$(PATH_CURRENT)/config/config.dist.yml"; \
	if [[ -f $$CONFIG ]]; then \
		$(call print,Configs replace,${COLOR_BLU}); \
		cp -R "$(PATH_CURRENT)/config/config.dist.yml" "${PATH_OUTPUT_BIN}/config.yml"; \
	fi

dependency: ## Download all dependencies
	$(call @print,Download all dependencies,${COLOR_BLU})
	@ERR=0 && \
	go mod download && \
	ERR=$$? && \
	if [ $$ERR != 0 ]; then \
		exit $$ERR; \
	fi

generate: ## File generation
	$(call @print,Generate files,${COLOR_BLU})
	@go generate ./...

tools-lint-ci: ## Install tools for golangci-lint
	@echo "Install golangci-lint"; \
    GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

tools-lint: ## Install tools for golint
	@export GO111MODULE=off && \
	export GOFLAGS="" && \
	echo "Install golint"; \
	go get -u -t golang.org/x/lint/golint

tools-format: ## Install tools for format
	@export GO111MODULE=off && \
	export GOFLAGS="" && \
    echo "Install goimports"; \
    go get -u -t golang.org/x/tools/cmd/goimports; \

tools-test: ## Install tools for test
	@export GO111MODULE=off && \
	export GOFLAGS="" && \
	echo "Install go-junit-report"; \
	go get -u github.com/jstemmer/go-junit-report; \
	echo "Install mockery"; \
	go get -u github.com/vektra/mockery/.../

.PHONY: test test-unit test-with-coverage test-race test-integration
##@ Test
test: test-unit test-with-coverage test-race ## Run test and show coverage
test-unit:
	$(call @print,Run unit tests,${COLOR_BLU})
	@go test -v ./...
test-with-coverage:
	$(call @print,Run unit tests with coverage,${COLOR_BLU})
	@go test -cover ./...
test-race:
	$(call @print,Run race detection,${COLOR_BLU})
	@go test -race  -count 100 ./...
test-integration: ## Launch integration tests
	$(call @print,Run integration tests,${COLOR_BLU})
	@docker-compose down && docker-compose build && docker-compose up --force-recreate

.PHONY: lint lint-format lint-style lint-ci lint-gosec
##@ Linter
lint: lint-format lint-ci lint-style ## Run code analysis (Check formatting, code style and check code with GolangCI-Lint )
lint-format: ## Check formatting
	$(call @print,Check formatting,${COLOR_BLU})
	@export GO111MODULE="off" && \
	export GOFLAGS="" && \
	errors=$$(gofmt -l -d $$(go list -f "{{ .Dir }}" ./...)); \
	if [[ "$${errors}" != "" ]]; then echo "$${errors}"; exit 1; fi
lint-ci: ## Check code with GolangCI-Lint
	$(call @print,Check code with GolangCI-Lint,${COLOR_BLU})
	@golangci-lint run

.PHONY: help
help:
	@awk 'BEGIN {RS = ""; FS="\n"; printf "\nVariables:\n"; } /^##\?.*\n[a-zA-Z_-]+/ { sub(/\?=.*/, "", $$2); printf "  \033[36m%-15s\033[0m %s\n", $$2, substr($$1, 5) }' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
