APP := td

.PHONY: test build

test:
	GOCACHE=$(PWD)/.cache/go-build GOMODCACHE=$(PWD)/.cache/gomod go test ./...

build:
	GOCACHE=$(PWD)/.cache/go-build GOMODCACHE=$(PWD)/.cache/gomod go build -o bin/$(APP) ./cmd/td
