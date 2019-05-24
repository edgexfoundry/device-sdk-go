.PHONY: build test clean docker

GO=CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=example/cmd/device-simple/device-simple
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-sdk-go.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) install -tags=safe

example/cmd/device-simple/device-simple:
	$(GO) build $(GOFLAGS) -o $@ ./example/cmd/device-simple

docker:
	docker build \
		-f example/cmd/device-simple/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-device-sdk-simple:$(GIT_SHA) \
		-t edgexfoundry/docker-device-sdk-simple:$(VERSION)-dev \
		.

test:
	$(GO) vet ./...
	gofmt -l .
	$(GO) test -coverprofile=coverage.out ./...

clean:
	rm -f $(MICROSERVICES)
