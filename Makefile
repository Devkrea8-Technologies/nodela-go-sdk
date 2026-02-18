.PHONY: help test lint fmt vet build install clean changeset version install-tools

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests
	go test -v ./tests -tags=integration

coverage: test ## Show test coverage
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run --timeout=5m

fmt: ## Format code
	gofmt -w -s .
	go mod tidy

vet: ## Run go vet
	go vet ./...

build: ## Build the project
	go build -v ./...

install: ## Install dependencies
	go mod download
	go mod verify

clean: ## Clean build artifacts
	go clean
	rm -f coverage.out

install-tools: ## Install development tools (changie)
	go install github.com/miniscruff/changie@latest

changeset: ## Create a new changeset describing your change
	changie new

version: ## Prepare a release - batches changesets and updates CHANGELOG.md (usage: make version VERSION=v1.2.3)
	@if [ -z "$(VERSION)" ]; then echo "Usage: make version VERSION=v1.2.3"; exit 1; fi
	changie batch $(VERSION)
	changie merge
	@echo ""
	@echo "Changelog updated for $(VERSION). Next steps:"
	@echo "  git add CHANGELOG.md .changes/"
	@echo "  git commit -m \"chore: prepare release $(VERSION)\""
	@echo "  git tag $(VERSION)"
	@echo "  git push && git push --tags"

.DEFAULT_GOAL := help