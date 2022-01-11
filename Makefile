.PHONY: build test clean docker unittest lint

ARCH=$(shell uname -m)
GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=example/cmd/device-simple/device-simple
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
DOCKER_TAG=$(VERSION)-dev

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-sdk-go/v2.Version=$(VERSION)"
GOTESTFLAGS?=-race

GIT_SHA=$(shell git rev-parse HEAD)

tidy:
	go mod tidy

build: $(MICROSERVICES)
	$(GOCGO) install -tags=safe

example/cmd/device-simple/device-simple:
	$(GOCGO) build $(GOFLAGS) -o $@ ./example/cmd/device-simple

docker:
	docker build \
		-f example/cmd/device-simple/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/device-simple:$(GIT_SHA) \
		-t edgexfoundry/device-simple:$(DOCKER_TAG) \
		.

unittest:
	GO111MODULE=on go test $(GOTESTFLAGS) -coverprofile=coverage.out ./...

lint:
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run\n  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.42.1"
	@if [ "z${ARCH}" = "zx86_64" ] && which golangci-lint >/dev/null ; then golangci-lint run --config .golangci.yml ; else echo "WARNING: Linting skipped (not on x86_64 or linter not installed)"; fi

test: unittest lint
	GO111MODULE=on go vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]
	./bin/test-attribution-txt.sh

clean:
	rm -f $(MICROSERVICES)

vendor:
	$(GO) mod vendor
