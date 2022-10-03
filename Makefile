.PHONY: build test clean docker unittest lint

ARCH=$(shell uname -m)
GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go


# see https://shibumi.dev/posts/hardening-executables
CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2"
CGO_CFLAGS="-O2 -pipe -fno-plt"
CGO_CXXFLAGS="-O2 -pipe -fno-plt"
CGO_LDFLAGS="-Wl,-O1,–sort-common,–as-needed,-z,relro,-z,now"

MICROSERVICES=example/cmd/device-simple/device-simple
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
DOCKER_TAG=$(VERSION)-dev

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-sdk-go/v2.Version=$(VERSION)" -trimpath -mod=readonly
CGOFLAGS=-ldflags "-linkmode=external -X github.com/edgexfoundry/device-sdk-go/v2.Version=$(VERSION)" -trimpath -mod=readonly -buildmode=pie
GOTESTFLAGS?=-race

GIT_SHA=$(shell git rev-parse HEAD)


build: $(MICROSERVICES)
	$(GOCGO) install -tags=safe

tidy:
	go mod tidy

example/cmd/device-simple/device-simple:
	$(GOCGO) build $(CGOFLAGS) -o $@ ./example/cmd/device-simple

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
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run\n  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.46.2"
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
