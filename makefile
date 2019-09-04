.PHONY: build test clean docker

GO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=examples/simple-filter-xml/simple-filter-xml examples/simple-cbor-filter/simple-cbor-filter examples/simple-filter-xml-mqtt/simple-filter-xml-mqtt examples/simple-filter-xml-post/simple-filter-xml-post examples/advanced-filter-convert-publish/advanced-filter-convert-publish examples/advanced-target-type/advanced-target-type
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/app-functions-sdk-go/internal.SDKVersion=$(VERSION) -X github.com/edgexfoundry/app-functions-sdk-go/internal.ApplicationVersion=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

examples/simple-filter-xml/simple-filter-xml:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml

examples/simple-cbor-filter/simple-cbor-filter:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-cbor-filter

examples/simple-filter-xml-mqtt/simple-filter-xml-mqtt:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml-mqtt

examples/simple-filter-xml-post/simple-filter-xml-post:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml-post

examples/advanced-filter-convert-publish/advanced-filter-convert-publish:
	$(GO) build $(GOFLAGS) -o $@ ./examples/advanced-filter-convert-publish

examples/advanced-target-type/advanced-target-type:
	$(GO) build $(GOFLAGS) -o $@ ./examples/advanced-target-type

docker:
	docker build \
		-f examples/simple-filter-xml/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(GIT_SHA) \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(VERSION)-dev \
		.

test:
	$(GO) test ./... -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./bin/test-go-mod-tidy.sh

clean:
	rm -f $(MICROSERVICES)