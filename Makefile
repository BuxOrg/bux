# Set the custom build tags
GO_BUILD_TAGS=json1

# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Not defined? Use default repo name which is the application
ifeq ($(REPO_NAME),)
	REPO_NAME="bux"
endif

## Not defined? Use default repo owner
ifeq ($(REPO_OWNER),)
	REPO_OWNER="BuxOrg"
endif

.PHONY: all
all: ## Runs multiple commands
	@$(MAKE) test-all-db

.PHONY: clean
clean: ## Remove previous builds and any cached data
	@echo "cleaning local cache..."
	@go clean -cache -testcache -i -r
	@$(MAKE) clean-mods
	@test $(DISTRIBUTIONS_DIR)
	@if [ -d $(DISTRIBUTIONS_DIR) ]; then rm -r $(DISTRIBUTIONS_DIR); fi

.PHONY: install-all-contributors
install-all-contributors: ## Installs all contributors locally
	@echo "installing all-contributors cli tool..."
	@yarn global add all-contributors-cli

.PHONY: release
release:: ## Runs common.release then runs godocs
	@$(MAKE) godocs

.PHONY: test-all-db
test-all-db: ## Runs all tests including embedded database tests
	@echo "running all tests including embedded database tests..."
	@go test ./... -v -tags="$(GO_BUILD_TAGS) database_tests"

.PHONY: test-all-db-ci
test-all-db-ci: ## Runs all tests including embedded database tests (CI)
	@echo "running all tests including embedded database tests..."
	@go test ./... -coverprofile=coverage.txt -covermode=atomic -tags="$(GO_BUILD_TAGS) database_tests"

.PHONE: update-contributors
update-contributors: ## Regenerates the contributors html/list
	@echo "generating contributor html..."
	@all-contributors generate
