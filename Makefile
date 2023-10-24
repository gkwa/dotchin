BIN := dotchin

GOPATH := $(shell go env GOPATH)
ifeq ($(OS),Windows_NT)
    GO_FILES := $(shell dir /S /B *.go)
    GO_DEPS := $(shell dir /S /B go.mod go.sum)
    CLEAN_CMD := del
else
    GO_FILES := $(shell find . -name '*.go')
    GO_DEPS := $(shell find . -name go.mod -o -name go.sum)
    CLEAN_CMD := rm -f
endif

$(BIN): $(GO_FILES) $(GO_DEPS)
	$(MAKE) pretty
	go vet ./...
	go build -o $(BIN) cmd/main.go

test: $(BIN)
	./$(BIN) --verbose
.PHONY: test

pretty: $(GO_FILES)
	gofumpt -w $^
.PHONY: pretty

install: $(GOPATH)/bin/$(BIN)
.PHONY: install

$(GOPATH)/bin/$(BIN): $(BIN)
	mv $(BIN) $(GOPATH)/bin/$(BIN)

clean:
	rm -f $(BIN)
.PHONY: clean
