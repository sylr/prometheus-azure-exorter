GIT_DESCRIBE ?= $(shell git describe --tags --dirty --broken || git rev-parse --short HEAD)

build:
	go build -ldflags "-X main.version=$(GIT_DESCRIBE)"

test:
	go test -ldflags "-X main.version=$(GIT_DESCRIBE)" ./...

install:
	go install -ldflags "-X main.version=$(GIT_DESCRIBE)"

go-mod-download:
	go mod download

go-mod-tidy:
	go mod tidy

go-dep-upgrade-minor:
	go get -u=patch ./...

go-dep-upgrade-major:
	go get -u ./...

verify-go-mod: go-mod-download
	git diff --quiet go.mod go.sum || git diff --exit-code go.mod go.sum

.PHONY: build test install go-mod-download go-mod-tidy verify-go-mod
