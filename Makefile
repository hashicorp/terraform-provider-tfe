TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=tfe
VERSION_FILE ?= version/VERSION
CHANGELOG_FILE = CHANGELOG.md

default: terraform-provider-tfe

build: fmtcheck
	go install

terraform-provider-tfe: fmtcheck
	@go build -o terraform-provider-tfe

# Run unit tests
test: fmtcheck
	go test -v $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test ./tfe -v -timeout 60m -sweep=prod

# Run acceptance tests
testacc: fmtcheck
	TF_ACC=1 TF_LOG_SDK_PROTO=OFF go test $(TEST) -v $(TESTARGS) -timeout 15m

# This rule creates a terraform CLI config file to override the tfe provider to point to the latest
# build in the current directory. The output of devoverride.sh is an export statement that
# overrides the CLI config to use this build.
devoverride: terraform-provider-tfe
	@sh -c "'$(CURDIR)/scripts/devoverride.sh'"

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

lint:
	@golangci-lint run ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "golangci-lint found some code style issues."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

exmplcheck:
	@./scripts/validate-example-presence.sh
	@./scripts/validate-import-example-presence.sh 
	@./scripts/validate-schema-descriptions.sh
	@./scripts/validate-examples.sh

generate:
	@RESOURCE="$(RESOURCE)" ./scripts/generate-docs.sh

# Release targets, used by release automation
prepare-release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is not set"; exit 1; fi
	@if [ -z "$(CHANGIE_VERSION)" ]; then echo "CHANGIE_VERSION is not set"; exit 1; fi
	
	@echo $(VERSION) > $(VERSION_FILE)
	@echo "Updated $(VERSION_FILE) to $(VERSION)"
	@sed -i.bak -E 's/~> [0-9]+\.[0-9]+\.[0-9]+/~> $(VERSION)/g' README.md && rm README.md.bak
	@npx -q changie@$(CHANGIE_VERSION) batch $(VERSION)
	@npx -q changie@$(CHANGIE_VERSION) merge

# Sanity check on version/VERSION, prepend the "Unreleased" section to the changelog,
# and overwrite version/VERSION with a dev label. The CHANGELOG will be overwritten by
# "changie merge" on the next prepare-release.
cleanup-release:
	@if [ -z "$(DEV_VERSION)" ]; then echo "DEV_VERSION is not set"; exit 1; fi
	@if ! git rev-parse --verify --quiet refs/tags/v$$(cat $(VERSION_FILE)) >/dev/null 2>&1; then echo "Latest version $$(cat $(VERSION_FILE)) has not been released"; exit 1; fi

	@TMP_FILE=$$(mktemp) ; \
		echo "## Unreleased" > "$$TMP_FILE" ; \
		echo "" >> "$$TMP_FILE" ; \
		echo "This file will be populated by automation before release. Below are the previously released changes:" >> "$$TMP_FILE" ; \
		echo "" >> "$$TMP_FILE" ; \
		cat $(CHANGELOG_FILE) >> "$$TMP_FILE" ; \
		mv "$$TMP_FILE" $(CHANGELOG_FILE) ;
	
	@echo "Release cleanup finished, version is now $(DEV_VERSION)"
	@echo $(DEV_VERSION) > $(VERSION_FILE)

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile sweep generate exmplcheck prepare-release cleanup-release