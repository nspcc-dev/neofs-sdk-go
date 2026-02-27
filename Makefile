#!/usr/bin/make -f

.PHONY: test test-full dep lint cover format modernize help

# Run tests
test:
	@go test ./... -cover

# Run tests, plus a docker tests with AIO
test-full:
	@go test -tags aiotest ./... -cover

# Pull go dependencies
dep:
	@printf "⇒ Download requirements: "
	@CGO_ENABLED=0 \
	go mod download && echo OK
	@printf "⇒ Tidy requirements: "
	@CGO_ENABLED=0 \
	go mod tidy -v && echo OK

.golangci.yml:
	wget -O $@ https://github.com/nspcc-dev/.github/raw/master/.golangci.yml

# Run linters
lint: .golangci.yml
	@golangci-lint --timeout=5m run

# Run tests with race detection and produce coverage output
cover:
	@go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -html=coverage.txt -o coverage.html

# Reformat code
format:
	@echo "⇒ Processing gofmt check"
	@gofmt -s -w ./
	@echo "⇒ Processing goimports check"
	@goimports -w ./

# Prettify code
modernize:
	@echo "⇒ Processing modernize check"
	@go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest -fix ./...


# Show this help prompt
help:
	@echo '  Usage:'
	@echo ''
	@echo '    make <target>'
	@echo ''
	@echo '  Targets:'
	@echo ''
	@awk '/^#/{ comment = substr($$0,3) } comment && /^[a-zA-Z][a-zA-Z0-9_-]+ ?:/{ print "   ", $$1, comment }' $(MAKEFILE_LIST) | column -t -s ':' | grep -v 'IGNORE' | sort -u
