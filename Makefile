FUZZ_TIME ?= 30s

.PHONY: all deps deps-ci test test-race fuzz fmt lint vuln verify check

all: fmt lint test

deps: deps-ci

deps-ci:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

test:
	go test ./...

test-race:
	go test -race ./...

fuzz:
	go test -run='^$$' -fuzz='^FuzzUnmarshalBinary$$' -fuzztime="$(FUZZ_TIME)" .

fmt:
	gofmt -w *.go example/*.go

lint:
	go vet ./...
	staticcheck ./...

vuln:
	govulncheck ./...

verify:
	go mod download
	go mod verify
	go mod tidy -diff

check: verify fmt lint vuln test-race
