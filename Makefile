SHELL        := $(shell which bash)

# Tooling is installed into BIN_DIR via `go run <tool>@<version>` so versions are
# pinned without network-piped install scripts.
LINTER       := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
TESTRUNNER   := go run gotest.tools/gotestsum@v1.13.0
ROOT_DIR     := $(shell git rev-parse --show-toplevel)
NO_COLOR     :=\033[0m
ATTN_COLOR   :=\033[33;01m

## EOF define block

.PHONY: all
all: fmt test lint cover

.PHONY: deps
deps:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go mod download

.PHONY: tidy
tidy:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go mod tidy

.PHONY: fmt
fmt:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gofmt -w $(ROOT_DIR)

.PHONY: test
test:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@CGO_ENABLED=0 $(TESTRUNNER) --format short-verbose -- -count=1 ./...

.PHONY: cover
cover:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@CGO_ENABLED=0 go test -count=1 -cover ./...

.PHONY: integration
integration:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@test -n "$(ABS_BASE_URL)" || { echo "set ABS_BASE_URL to a running audiobookshelf server"; exit 1; }
	@CGO_ENABLED=0 go test -tags=integration -count=1 .

.PHONY: vet
vet:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go vet ./...

.PHONY: lint
lint:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@CGO_ENABLED=0 $(LINTER) run ./...

.PHONY: clean
clean:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@rm -rf $(BIN_DIR)
	@rm -rf $(REL_DIR)
	@go clean

$(REL_DIR):
	@echo -e "$(ATTN_COLOR)==> create REL_DIR $(REL_DIR) $(NO_COLOR)"
	@mkdir -p $(REL_DIR)

$(BIN_DIR):
	@echo -e "$(ATTN_COLOR)==> create BIN_DIR $(BIN_DIR) $(NO_COLOR)"
	@mkdir -p $(BIN_DIR)
