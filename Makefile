# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show list of make targets and their description
	@grep -E '^[/%.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL:= help

.PHONY: setup
setup: ## Run setup scripts to prepare development environment
	@scripts/setup.sh

.PHONY: gen
gen: ## Show gen.sh help
	@scripts/gen.sh
gen/%: ## Generate artifacts that defined by '%', e.g: 'make gen/proto` will trigger ./scripts/gen.sh proto
	@scripts/gen.sh $*

.PHONY: lint
lint: ## Run linter
	spkit lint

.PHONY: test
test: ## Generate mock and run all test
	make gen/all
	spkit test

.PHONY: build
build: ## Build all main packages
	@scripts/build.sh

.PHONY: clean
clean: ## Clean project dir, remove build artifacts and logs
	spkit clean .

all: clean setup build ## Clean, setup, generate and then build all the binaries.

