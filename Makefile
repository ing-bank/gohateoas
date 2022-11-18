MAKEFLAGS := --no-print-directory --silent

default: help

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z\._-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

t: test
test: fmt ## Run unit tests, alias: t
	go test ./... -timeout=30s -parallel=8

fmt: ## Format go code
	@go mod tidy
	@go fmt ./...
