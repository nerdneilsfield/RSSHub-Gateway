projectname?=rsshub-gateway

VERSION ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(DATE) -X main.gitCommit=$(COMMIT)

default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build golang binary
	@go build -ldflags "$(LDFLAGS)" -o $(projectname) ./

.PHONY: install
install: ## install golang binary
	@go install -ldflags "$(LDFLAGS)"

.PHONY: run
run: ## run the app
	@go run -ldflags "$(LDFLAGS)" ./ serve -c config.example.yaml

.PHONY: bootstrap
bootstrap: ## install build deps
	@go mod download

.PHONY: test
test: ## run tests
	go test ./...

.PHONY: clean
clean: ## clean up environment
	@rm -rf coverage.out dist/ $(projectname)

.PHONY: cover
cover: ## display test coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: fmt
fmt: ## format go files
	gofmt -w .

.PHONY: lint
lint: ## lint go files
	golangci-lint run -c .golangci.yml

.PHONY: release-test
release-test: ## test release
	goreleaser release --snapshot --clean --skip-publish

# .PHONY: pre-commit
# pre-commit:	## run pre-commit hooks
# 	pre-commit run --all-files
