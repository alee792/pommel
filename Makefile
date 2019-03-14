SHELL := /bin/bash
mod := github.com/alee792/pommel
pkg := pommel
version := $(shell git describe --tags)
$(info $(pkg) - $(version))

export GO111MODULE=on

.PHONY: run
run: ## Run the cmd
	@go run $(mod)/cmd/$(pkg)

.PHONY: bin
bin: ## Create binary
	@go build $(mod)/cmd/$(pkg)

.PHONY: test
test: ## Test packages
	@go test -race -count 1 -coverprofile /tmp/c.out $(mod)/internal/...

.PHONY: badges
badges: ## Create badges
	@gopherbadger -md=README.md -png=false

.PHONY: tidy
tidy: ## Tidy go.mod
	@go mod tidy

.PHONY: clean
clean: ## Remove artifacts
	@rm $(pkg) 2>/dev/null || true

.PHONY: push
push: test ## Push to remote
	git push

.PHONY: release 
release:  test ## Release a tag
ifeq ($(v),)
	$(eval p := $(shell git describe --tags --abbrev=0 | sed -Ee 's/^v|-.*//'))
ifeq ($(bump), major)
	$(eval f := 1)
else ifeq ($(bump), minor)
	$(eval f := 2)
else
	$(eval f := 3)
endif
	$(eval new := v$(shell echo $(p) | awk -F. -v OFS=. -v f=$(f) '{ $$f++ } 1'))
else
	$(eval new := $(v))
endif

ifeq ($(l), t)
	git tag latest
endif
	echo $(new)
	git tag "$(new)"
	git push origin "$(new)"

.PHONY: help
help: ## List targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
