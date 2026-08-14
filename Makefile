all: lint build test

build:
	go build ./...

install:
	./scripts/make-install.sh

lint: fmt vet misspell

fmt:
	./scripts/gofmt.sh

vet:
	go vet ./check/... ./cmd/...

test:
	go test -cover ./...

badge: install
	go run ./cmd/goreportcard-cli -svg goreportcard.svg

misspell:
	@[ -x "$(shell which misspell)" ] || go install github.com/client9/misspell/cmd/misspell
	find . -name '*.go' -not -path './vendor/*' -not -path './check/testdata/*' | xargs misspell -error
